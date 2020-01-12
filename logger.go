package flog

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/buexplain/go-flog/constant"
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
	handlers []HandlerInterface
	//额外日志信息处理器集合
	extras []ExtraInterface
	//日志收集齐器关闭状态
	closed chan struct{}
	//异步日志队列
	queue chan *Record
	//异步日志队列关闭状态
	queueClosed chan struct{}
	//异步日志队列处理go程退出时候的等待超时时间
	timeout time.Duration
	//关闭锁
	lock *sync.Mutex
}

func New(channel string, handler HandlerInterface, extra ...ExtraInterface) *Logger {
	tmp := new(Logger)
	tmp.channel = channel
	tmp.handlers = []HandlerInterface{handler}
	tmp.extras = make([]ExtraInterface, 0, len(extra))
	tmp.extras = append(tmp.extras, extra...)
	tmp.closed = make(chan struct{})
	tmp.queue = nil
	tmp.timeout = 2 * time.Second
	tmp.lock = new(sync.Mutex)
	return tmp
}

func (this *Logger) Async(capacity int) {
	if this.queue == nil {
		this.queue = make(chan *Record, capacity)
		this.queueClosed = make(chan struct{})
		//开启日志异步写入go程
		go this.goF()
	}
}

//日志异步写入go程
func (this *Logger) goF() {
	defer func() {
		if a := recover(); a != nil {
			//记录错误栈
			var message string
			message = fmt.Sprintf("Logger %s uncaught panic: %s", this.channel, debug.Stack())
			libLog.Println(message)
			//重启一条go程
			go this.goF()
		} else {
			//正常退出，发出日志队列已经关闭信号
			close(this.queueClosed)
		}
	}()
	for {
		select {
		case record := <-this.queue:
			//调度日志
			this.dispatch(record)
			break
		case <-this.closed:
			//收到日志关闭信号
			//调度完日志队列中剩余的日志
			for {
				for i := len(this.queue); i > 0; i-- {
					this.dispatch(<-this.queue)
				}
				if len(this.queue) == 0 {
					break
				}
			}
			//再次调度日志队列中剩余的日志，直到超时退出
			for {
				select {
				case record := <-this.queue:
					this.dispatch(record)
				case <-time.After(this.timeout):
					return
				}
			}
		}
	}
}

func (this *Logger) Close(timeout ...time.Duration) error {
	//获取锁
	this.lock.Lock()
	defer this.lock.Unlock()

	//判断是否关闭
	select {
	case <-this.closed:
		return nil
	default:
		break
	}
	//发出日志关闭信号
	close(this.closed)

	//异步日志，清空日志队列中的日志
	if this.queue != nil {
		if len(timeout) > 0 {
			this.timeout = timeout[0]
		}
		//等待日志队列清空
		<-this.queueClosed
		//关闭日志队列
		close(this.queue)
	}

	//关闭各个日志处理器
	err := bytes.Buffer{}
	for _, v := range this.handlers {
		if e := v.Close(); e != nil {
			err.WriteString(e.Error())
			err.WriteString(constant.EOF)
		}
	}
	//返回错误
	if err.Len() > 0 {
		return errors.New(err.String())
	}
	return nil
}

func (this *Logger) GetChannel() string {
	return this.channel
}

func (this *Logger) PushHandler(handler HandlerInterface) *Logger {
	this.handlers = append(this.handlers, handler)
	return this
}

func (this *Logger) PopHandler() HandlerInterface {
	if len(this.handlers) == 0 {
		return nil
	}
	tmp := this.handlers[len(this.handlers)-1]
	this.handlers = this.handlers[0 : len(this.handlers)-1]
	return tmp
}

func (this *Logger) GetHandlers() []HandlerInterface {
	return this.handlers
}

func (this *Logger) PushExtra(extra ExtraInterface) *Logger {
	this.extras = append(this.extras, extra)
	return this
}

func (this *Logger) PopExtra() ExtraInterface {
	if len(this.extras) == 0 {
		return nil
	}
	tmp := this.extras[len(this.extras)-1]
	this.extras = this.extras[0 : len(this.extras)-1]
	return tmp
}

func (this *Logger) GetExtras(extra ExtraInterface) []ExtraInterface {
	return this.extras
}

func (this *Logger) AddRecord(level Level, format bool, message string, context ...interface{}) {
	//判断日志收集齐器状态
	select {
	case <-this.closed:
		//日志收集齐器处于关闭状态，不再收集日志
		return
	default:
		//继续收集日志
		break
	}

	//判断是否有日志处理器可以处理当前level的日志
	isHandling := false
	for _, v := range this.handlers {
		if v.IsHandling(level) {
			isHandling = true
			break
		}
	}
	if !isHandling {
		return
	}

	//新建一个日志载体对象
	record := NewRecord()
	record.Channel = this.channel
	record.Level = level
	record.LevelName = GetNameByLevel(level)
	if format {
		record.Message = fmt.Sprintf(message, context...)
	} else {
		record.Message = message
		if len(context) > 0 {
			record.Context = context
		}
	}

	//给日志对象添加额外信息
	for _, v := range this.extras {
		v.Processor(record)
	}

	//调度日志
	if this.queue == nil {
		//同步调度
		this.dispatch(record)
	} else {
		//异步抛入日志队列
		this.queue <- record
	}
}

func (this *Logger) dispatch(record *Record) {
	for _, v := range this.handlers {
		if v.IsHandling(record.Level) {
			if v.Handle(record) {
				break
			}
		}
	}
}

/**
 * 紧急情况：系统无法使用
 */
func (this *Logger) Emergency(message string, context ...interface{}) {
	this.AddRecord(LEVEL_EMERGENCY, false, message, context...)
}

func (this *Logger) EmergencyF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_EMERGENCY, true, format, v...)
}

/**
 *警报：必须立即采取措施
 */
func (this *Logger) Alert(message string, context ...interface{}) {
	this.AddRecord(LEVEL_ALERT, false, message, context...)
}

func (this *Logger) AlertF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_ALERT, true, format, v...)
}

/**
 * 严重：危急情况
 */
func (this *Logger) Critical(message string, context ...interface{}) {
	this.AddRecord(LEVEL_CRITICAL, false, message, context...)
}

func (this *Logger) CriticalF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_CRITICAL, true, format, v...)
}

/**
 * 错误
 */
func (this *Logger) Error(message string, context ...interface{}) {
	this.AddRecord(LEVEL_ERROR, false, message, context...)
}

func (this *Logger) ErrorF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_ERROR, true, format, v...)
}

/**
 * 警告
 */
func (this *Logger) Warning(message string, context ...interface{}) {
	this.AddRecord(LEVEL_WARNING, false, message, context...)
}

func (this *Logger) WarningF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_WARNING, true, format, v...)
}

/**
 * 注意：正常但重要条件
 */
func (this *Logger) Notice(message string, context ...interface{}) {
	this.AddRecord(LEVEL_NOTICE, false, message, context...)
}

func (this *Logger) NoticeF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_NOTICE, true, format, v...)
}

/**
 * 信息
 */
func (this *Logger) Info(message string, context ...interface{}) {
	this.AddRecord(LEVEL_INFO, false, message, context...)
}

func (this *Logger) InfoF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_INFO, true, format, v...)
}

/**
 * 调试
 */
func (this *Logger) Debug(message string, context ...interface{}) {
	this.AddRecord(LEVEL_DEBUG, false, message, context...)
}

func (this *Logger) DebugF(format string, v ...interface{}) {
	this.AddRecord(LEVEL_DEBUG, true, format, v...)
}
