package handler

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/buexplain/go-flog"
	"github.com/buexplain/go-flog/constant"
	"io"
	"io/ioutil"
	libLog "log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

//文件日志处理器
type File struct {
	//日志等级
	level flog.Level
	//日志格式化处理器
	formatter flog.FormatterInterface
	//是否阻止进入下一个日志处理器
	bubble bool
	//日志写入路径
	path string
	//日志文件权限
	perm os.FileMode
	//单个日志文件最大写入大小，0则不做限制
	maxSize int64
	//日志文件写入锁
	lock *sync.Mutex
	//当前日志文件指针
	file *os.File
	//当前日志文件写入大小
	currentSize int64
	//今日结束时间
	todayEndUnix int64
	//写入缓冲区
	buffer *bufio.Writer
	//写入缓冲区关闭状态
	bufferClosed chan struct{}
	//缓冲区冲刷时间间隔
	flush time.Duration
	//日志写入接口
	w io.Writer
	//文件日志处理器关闭状态
	closed chan struct{}
}

func NewFile(level flog.Level, formatter flog.FormatterInterface, path string) *File {
	tmp := new(File)
	tmp.level = level
	tmp.formatter = formatter
	tmp.bubble = false
	tmp.setPath(path)
	tmp.perm = 0666
	tmp.maxSize = 256 << 20
	tmp.lock = new(sync.Mutex)
	tmp.buffer = nil
	tmp.closed = make(chan struct{})
	return tmp
}

func (this *File) SetBuffer(buffer int, flush time.Duration) {
	if this.buffer == nil {
		//初始化缓冲区
		this.buffer = bufio.NewWriterSize(nil, buffer)
		this.flush = flush
		this.bufferClosed = make(chan struct{})
		//开启定时冲刷缓冲区的go程
		go this.goF()
	}
}

//定时冲刷缓冲区
func (this *File) goF() {
	ticker := time.NewTicker(this.flush)
	defer func() {
		ticker.Stop()
		if a := recover(); a != nil {
			//记录错误栈
			var message string
			message = fmt.Sprintf("File handler uncaught panic: %s", debug.Stack())
			libLog.Println(message)
			//重启一条go程
			this.goF()
		} else {
			//正常退出go程
			close(this.bufferClosed)
		}
	}()
	for {
		select {
		case <-this.closed:
			return
		default:
			<-ticker.C
			this.lock.Lock()
			if err := this.buffer.Flush(); err != nil {
				libLog.Println(err)
			}
			this.lock.Unlock()
		}
	}
}

func (this *File) setPath(path string) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		libLog.Panicln(err)
	}
	path = filepath.ToSlash(path)
	if err := os.MkdirAll(path, os.ModeSetgid); err != nil {
		libLog.Panicln(err)
	}
	this.path = path
}

func (this *File) GetPath() string {
	return this.path
}

func (this *File) SetBubble(bubble bool) {
	this.bubble = bubble
}

func (this *File) SetPerm(perm os.FileMode) {
	this.perm = perm
}

func (this *File) SetMaxSize(maxSize int64) {
	if maxSize > 0 {
		this.maxSize = maxSize
	}
}

func (this *File) Close() error {
	//获取锁
	this.lock.Lock()
	//退出的时候释放锁
	defer this.lock.Unlock()

	//判断是否关闭
	select {
	case <-this.closed:
		return nil
	default:
		break
	}

	//发出关闭信号
	close(this.closed)

	err := bytes.Buffer{}

	//清空缓冲区
	if this.buffer != nil {
		//释放锁，让缓冲区冲刷go程有机会拿到锁
		this.lock.Unlock()
		//等待缓冲区冲刷go程结束
		<-this.bufferClosed
		//再次获取锁
		this.lock.Lock()
		//再次冲刷缓冲区
		if e := this.buffer.Flush(); e != nil {
			err.WriteString(e.Error())
			err.WriteString(constant.EOF)
		}
		//释放缓冲区
		this.buffer.Reset(nil)
		this.buffer = nil
	}

	//关闭文件指针
	if this.file != nil {
		if e := this.file.Close(); e != nil {
			err.WriteString(e.Error())
			err.WriteString(constant.EOF)
		}
	}

	//返回错误信息
	if err.Len() > 0 {
		return errors.New(err.String())
	}
	return nil
}

func (this *File) IsHandling(level flog.Level) bool {
	return level <= this.level
}

func (this *File) Handle(record *flog.Record) bool {
	//获取锁
	this.lock.Lock()
	defer this.lock.Unlock()

	//判断文件日志处理器处于关闭状态
	select {
	case <-this.closed:
		//文件日志处理器处于关闭状态，不再处理日志，强制返回false，让下一个日志handler继续处理日志信息
		return false
	default:
		//继续处理日志
		break
	}

	//当前日志文件指针不存在，或者当前日志文件已经写满，或者日期已经跳到了第二天，则创建新的日志文件指针
	if this.file == nil || this.currentSize > this.maxSize || record.Time.Unix() > this.todayEndUnix {
		//尝试关闭旧的日志文件指针
		if this.file != nil {
			//先尝试刷新缓冲区中的缓冲
			if this.buffer != nil {
				if err := this.buffer.Flush(); err != nil {
					//刷新失败，打印日志
					libLog.Println(err)
				}
			}
			//开始关闭旧的日志文件指针
			if err := this.file.Close(); err != nil {
				//关闭失败，打印日志
				libLog.Println(err)
				//强制返回false，让下一个日志handler继续处理日志信息
				return false
			} else {
				//关闭成功，设置为nil
				this.file = nil
			}
		}
		//寻找新的日志文件名
		var logName string
		var logNameIndex int
		logNameIndex, err := findLogNameIndex(this.path, record.Time.Format("2006-01-02"))
		if err != nil {
			//关闭失败，打印日志
			libLog.Println(err)
			//强制返回false，让下一个日志handler继续处理日志信息
			return false
		} else if logNameIndex == -1 {
			logName = filepath.Join(this.path, record.Time.Format("2006-01-02.log"))
		} else {
			logName = filepath.Join(this.path, record.Time.Format("2006-01-02")+fmt.Sprintf(".%d.log", logNameIndex))
		}
		//检查日志文件名
		if fi, err := os.Stat(logName); err == nil {
			//文件已经存在，判断是否超出写入大小
			if fi.Size() >= this.maxSize {
				//超出大小，生成一个新的文件名
				logName = filepath.Join(this.path, record.Time.Format("2006-01-02")+fmt.Sprintf(".%d.log", logNameIndex+1))
			}
		} else if !os.IsNotExist(err) {
			//不是文件不存在错误，调用标准库日志打印错误
			libLog.Println(err)
			//强制返回false，让下一个日志handler继续处理日志信息
			return false
		}
		//打开或创建日志文件
		if f, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, this.perm); err == nil {
			if fi, err := f.Stat(); err == nil {
				this.file = f
				this.currentSize = fi.Size()
				this.todayEndUnix = time.Date(record.Time.Year(), record.Time.Month(), record.Time.Day(), 23, 59, 59, 0, record.Time.Location()).Unix()
				//设置日志写入接口
				if this.buffer != nil {
					//开启日志缓冲，重置缓冲区
					this.buffer.Reset(this.file)
					this.w = this.buffer
				} else {
					this.w = this.file
				}
			} else {
				//错误，调用标准库日志打印错误
				libLog.Println(err)
				//强制返回false，让下一个日志handler继续处理日志信息
				return false
			}
		} else {
			//错误，调用标准库日志打印错误
			libLog.Println(err)
			//强制返回false，让下一个日志handler继续处理日志信息
			return false
		}
	}

	if n, err := this.formatter.Format(this.w, record); err == nil {
		this.currentSize += int64(n)
		return this.bubble
	} else {
		//错误，调用标准库日志打印错误
		libLog.Println(err)
		//强制返回false，让下一个日志handler继续处理日志信息
		return false
	}
}

func findLogNameIndex(path string, date string) (int, error) {
	fiArr, err := ioutil.ReadDir(path)
	if err != nil {
		return -1, err
	}
	reg := regexp.MustCompile(`^` + date + `([0-9\\.]*)\.log$`)
	index := -1
	for _, fi := range fiArr {
		subArr := reg.FindStringSubmatch(fi.Name())
		if len(subArr) == 0 {
			continue
		}
		subArr[1] = strings.TrimLeft(subArr[1], ".")
		if subArr[1] == "" {
			continue
		}
		tmp, err := strconv.Atoi(subArr[1])
		if err == nil && tmp > index && strconv.Itoa(tmp) == subArr[1] {
			index = tmp
		}
	}
	return index, nil
}
