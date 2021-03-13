package flog

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/buexplain/go-flog/contract"
	libLog "log"
	"runtime/debug"
	"sync"
	"time"
)

//日志收集齐器
type Logger struct {
	//渠道名称
	channel string
	//日志处理器集合
	handlers []contract.Handler
	//额外日志信息处理器集合
	extras []contract.Extra
	//日志收集齐器关闭状态
	closed chan struct{}
	//异步日志队列
	queue chan *contract.Record
	//异步日志队列关闭状态
	queueClosed chan struct{}
	//异步日志队列处理go程退出时候的等待超时时间
	timeout time.Duration
	//关闭锁
	lock *sync.Mutex
}

func New(channel string, handler contract.Handler, extra ...contract.Extra) *Logger {
	tmp := new(Logger)
	tmp.channel = channel
	tmp.handlers = []contract.Handler{}
	tmp.PushHandler(handler)
	tmp.extras = make([]contract.Extra, 0, len(extra))
	tmp.extras = append(tmp.extras, extra...)
	tmp.closed = make(chan struct{})
	tmp.queue = nil
	tmp.timeout = 2 * time.Second
	tmp.lock = new(sync.Mutex)
	return tmp
}

func (r *Logger) Async(capacity int) {
	if r.queue == nil {
		r.queue = make(chan *contract.Record, capacity)
		r.queueClosed = make(chan struct{})
		//开启日志异步写入go程
		go r.goF()
	}
}

//日志异步写入go程
func (r *Logger) goF() {
	defer func() {
		if a := recover(); a != nil {
			//记录错误栈
			var message string
			message = fmt.Sprintf("Logger %s uncaught panic: %s", r.channel, debug.Stack())
			libLog.Println(message)
			//重启一条go程
			go r.goF()
		} else {
			//正常退出，发出日志队列已经关闭信号
			close(r.queueClosed)
		}
	}()
	for {
		select {
		case record := <-r.queue:
			//调度日志
			r.dispatch(record)
			break
		case <-r.closed:
			//收到日志关闭信号
			//调度完日志队列中剩余的日志
			for {
				for i := len(r.queue); i > 0; i-- {
					r.dispatch(<-r.queue)
				}
				if len(r.queue) == 0 {
					break
				}
			}
			//再次调度日志队列中剩余的日志，直到超时退出
			for {
				select {
				case record := <-r.queue:
					r.dispatch(record)
				case <-time.After(r.timeout):
					return
				}
			}
		}
	}
}

func (r *Logger) Close(timeout ...time.Duration) error {
	//获取锁
	r.lock.Lock()
	defer r.lock.Unlock()

	//判断是否关闭
	select {
	case <-r.closed:
		return nil
	default:
		break
	}

	//设置超时
	if len(timeout) > 0 {
		r.timeout = timeout[0]
	}

	//发出日志关闭信号
	close(r.closed)

	//异步日志，清空日志队列中的日志
	if r.queue != nil {
		//等待日志队列清空
		<-r.queueClosed
		//关闭日志队列
		close(r.queue)
	}

	//关闭各个日志处理器
	bag := bytes.Buffer{}
	for _, v := range r.handlers {
		if e := v.Close(); e != nil {
			bag.WriteString(e.Error())
			bag.WriteByte('\n')
		}
	}

	//返回错误
	if bag.Len() > 0 {
		return errors.New(bag.String())
	}
	return nil
}

func (r *Logger) GetChannel() string {
	return r.channel
}

func (r *Logger) PushHandler(handler contract.Handler) *Logger {
	if handler != nil {
		r.handlers = append(r.handlers, handler)
	}
	return r
}

func (r *Logger) PopHandler() contract.Handler {
	if len(r.handlers) == 0 {
		return nil
	}
	tmp := r.handlers[len(r.handlers)-1]
	r.handlers = r.handlers[0 : len(r.handlers)-1]
	return tmp
}

func (r *Logger) GetHandlers() []contract.Handler {
	return r.handlers
}

func (r *Logger) PushExtra(extra contract.Extra) *Logger {
	r.extras = append(r.extras, extra)
	return r
}

func (r *Logger) PopExtra() contract.Extra {
	if len(r.extras) == 0 {
		return nil
	}
	tmp := r.extras[len(r.extras)-1]
	r.extras = r.extras[0 : len(r.extras)-1]
	return tmp
}

func (r *Logger) GetExtras() []contract.Extra {
	return r.extras
}

func (r *Logger) AddRecord(level contract.Level, format bool, message string, context ...interface{}) {
	//判断是否有日志处理器可以处理当前level的日志
	isHandling := false
	for _, v := range r.handlers {
		if v.IsHandling(level) {
			isHandling = true
			break
		}
	}
	if !isHandling {
		return
	}

	//新建一个日志载体对象
	record := contract.NewRecord()
	record.Channel = r.channel
	record.Level = contract.GetNameByLevel(level)
	if format {
		record.Message = fmt.Sprintf(message, context...)
	} else {
		record.Message = message
		if l := len(context); l > 0 {
			if l == 1 {
				record.Context = context[0]
			} else {
				record.Context = context
			}
		}
	}

	//给日志对象添加额外信息
	for _, v := range r.extras {
		v.Processor(record)
	}

	//判断日志收集齐器状态
	select {
	case <-r.closed:
		//日志收集齐器处于关闭状态，不再收集日志
		return
	default:
		//继续收集日志
		break
	}

	//调度日志
	if r.queue == nil {
		//同步调度
		r.dispatch(record)
	} else {
		//异步抛入日志队列
		r.queue <- record
	}
}

func (r *Logger) dispatch(record *contract.Record) {
	for _, v := range r.handlers {
		if v.IsHandling(contract.GetLevelByName(record.Level)) {
			if v.Handle(record) {
				break
			}
		}
	}
}

/**
 * 紧急情况：系统无法使用
 */
func (r *Logger) Emergency(message string, context ...interface{}) {
	r.AddRecord(contract.LevelEmergency, false, message, context...)
}

func (r *Logger) EmergencyF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelEmergency, true, format, v...)
}

/**
 *警报：必须立即采取措施
 */
func (r *Logger) Alert(message string, context ...interface{}) {
	r.AddRecord(contract.LevelAlert, false, message, context...)
}

func (r *Logger) AlertF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelAlert, true, format, v...)
}

/**
 * 严重：危急情况
 */
func (r *Logger) Critical(message string, context ...interface{}) {
	r.AddRecord(contract.LevelCritical, false, message, context...)
}

func (r *Logger) CriticalF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelCritical, true, format, v...)
}

/**
 * 错误
 */
func (r *Logger) Error(message string, context ...interface{}) {
	r.AddRecord(contract.LevelError, false, message, context...)
}

func (r *Logger) ErrorF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelError, true, format, v...)
}

/**
 * 警告
 */
func (r *Logger) Warning(message string, context ...interface{}) {
	r.AddRecord(contract.LevelWarning, false, message, context...)
}

func (r *Logger) WarningF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelWarning, true, format, v...)
}

/**
 * 注意：正常但重要条件
 */
func (r *Logger) Notice(message string, context ...interface{}) {
	r.AddRecord(contract.LevelNotice, false, message, context...)
}

func (r *Logger) NoticeF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelNotice, true, format, v...)
}

/**
 * 信息
 */
func (r *Logger) Info(message string, context ...interface{}) {
	r.AddRecord(contract.LevelInfo, false, message, context...)
}

func (r *Logger) InfoF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelInfo, true, format, v...)
}

/**
 * 调试
 */
func (r *Logger) Debug(message string, context ...interface{}) {
	r.AddRecord(contract.LevelDebug, false, message, context...)
}

func (r *Logger) DebugF(format string, v ...interface{}) {
	r.AddRecord(contract.LevelDebug, true, format, v...)
}
