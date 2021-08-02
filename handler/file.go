package handler

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/buexplain/go-flog/contract"
	"io"
	libLog "log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

// File 文件日志处理器
type File struct {
	//日志等级
	level contract.Level
	//日志格式化处理器
	formatter contract.Formatter
	//是否阻止进入下一个日志处理器
	bubble bool
	//日志写入路径
	path string
	//日志文件名字前缀
	prefix string
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
	//日志写入接口
	w io.Writer
	//处理器关闭锁
	closeLock *sync.Mutex
	//写入锁
	writeLock *sync.Mutex
	//文件日志处理器关闭状态
	closed chan struct{}
	//写入缓冲区
	buffer *bufio.Writer
	//写入缓冲区关闭状态
	bufferClosed chan struct{}
	//缓冲区冲刷时间间隔
	flush time.Duration
}

func NewFile(level contract.Level, formatter contract.Formatter, path string) *File {
	tmp := new(File)
	tmp.level = level
	tmp.formatter = formatter
	tmp.bubble = false
	tmp.setPath(path)
	tmp.perm = 0666
	tmp.maxSize = 256 << 20
	tmp.closeLock = new(sync.Mutex)
	tmp.writeLock = new(sync.Mutex)
	tmp.buffer = nil
	tmp.closed = make(chan struct{})
	return tmp
}

func (r *File) SetBuffer(size int, flush time.Duration) {
	if r.buffer == nil {
		r.buffer = bufio.NewWriterSize(nil, size)
		r.flush = flush
		r.bufferClosed = make(chan struct{})
		go r.goF()
	}
}

//定时冲刷缓冲区
func (r *File) goF() {
	ticker := time.NewTicker(r.flush)
	defer ticker.Stop()
	defer func() {
		if a := recover(); a != nil {
			libLog.Println(fmt.Sprintf("File handler uncaught panic: %s", debug.Stack()))
			r.goF()
		} else {
			close(r.bufferClosed)
		}
	}()
	for {
		select {
		case <-r.closed:
			r.writeLock.Lock()
			if err := r.buffer.Flush(); err != nil {
				r.writeLock.Unlock()
				libLog.Println(err)
			} else {
				r.writeLock.Unlock()
			}
			return
		case <-ticker.C:
			r.writeLock.Lock()
			if err := r.buffer.Flush(); err != nil {
				r.writeLock.Unlock()
				libLog.Println(err)
			} else {
				r.writeLock.Unlock()
			}
		}
	}
}

func (r *File) setPath(path string) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		libLog.Panicln(err)
	}
	path = filepath.ToSlash(path)
	if err := os.MkdirAll(path, os.ModeSetgid); err != nil {
		libLog.Panicln(err)
	}
	r.path = path
}

func (r *File) GetPath() string {
	return r.path
}

func (r *File) SetPrefix(prefix string)  {
	r.prefix = prefix
}

func (r *File) SetBubble(bubble bool) {
	r.bubble = bubble
}

func (r *File) SetPerm(perm os.FileMode) {
	r.perm = perm
}

func (r *File) SetMaxSize(maxSize int64) {
	if maxSize > 0 {
		r.maxSize = maxSize
	}
}

func (r *File) Close() error {
	r.closeLock.Lock()
	defer r.closeLock.Unlock()
	select {
	case <-r.closed:
		return nil
	default:
		break
	}
	close(r.closed)
	bag := bytes.Buffer{}
	if r.buffer != nil {
		<-r.bufferClosed
		r.buffer.Reset(nil)
		r.buffer = nil
	}
	r.writeLock.Lock()
	defer r.writeLock.Unlock()
	if r.file != nil {
		if e := r.file.Close(); e != nil {
			bag.WriteString(e.Error())
			bag.WriteByte('\n')
		}
	}
	if bag.Len() > 0 {
		return errors.New(bag.String())
	}
	return nil
}

func (r *File) IsHandling(level contract.Level) bool {
	return level <= r.level
}

//找到日期下最后一个日志文件的索引值
func (r *File) findLogNameLastIndex(date string) (index int, err error) {
	var entries []os.DirEntry
	entries, err = os.ReadDir(r.path)
	if err != nil {
		return -1, err
	}
	var reg *regexp.Regexp
	if r.prefix == "" {
		reg = regexp.MustCompile(`^` + date + `[\.]*([0-9]*)\.log$`)
	}else {
		reg = regexp.MustCompile(`^` + r.prefix+ "-"+ date + `[\.]*([0-9]*)\.log$`)
	}
	index = -1
	var currentIndex int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		sub := reg.FindStringSubmatch(entry.Name())
		if len(sub) != 2 || sub[1] == "" {
			continue
		}
		currentIndex, err = strconv.Atoi(sub[1])
		if err == nil && currentIndex > index {
			index = currentIndex
		}
	}
	return index, nil
}

// scanLogName 寻找新的日志文件名
func (r *File) scanLogName(t time.Time) (name string, err error) {
	var index int
	index, err = r.findLogNameLastIndex(t.Format("2006-01-02"))
	if err != nil {
		return "", err
	}
	if index == -1 {
		name = filepath.Join(r.path, t.Format("2006-01-02.log"))
	} else {
		name = filepath.Join(r.path, t.Format("2006-01-02")+fmt.Sprintf(".%d.log", index))
	}
	//检查日志文件名
	var fi os.FileInfo
	fi, err = os.Stat(name)
	//文件已经存在
	if err == nil {
		if r.maxSize > 0 && fi.Size() >= r.maxSize {
			//超出写入大小限制，生成一个新的文件名
			name = filepath.Join(r.path, t.Format("2006-01-02")+fmt.Sprintf(".%d.log", index+1))
		}
		return name, nil
	}
	//出错了，并且不是：文件不存在错误
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	//文件不存在，可以使用该名字
	return name, nil
}

// closeFile 关闭当前的日志文件
func (r *File) closeFile() error {
	if r.file == nil {
		return nil
	}
	//检查缓冲区，并刷新缓冲区
	if r.buffer != nil {
		if err := r.buffer.Flush(); err != nil {
			return err
		}
	}
	//关闭旧的日志文件指针
	if err := r.file.Close(); err != nil {
		return err
	}
	//关闭成功，设置为nil
	r.file = nil
	return nil
}

// openFile 打开或创建日志文件
func (r *File) openFile(logName string, t time.Time) error {
	//打开日志文件
	f, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, r.perm)
	if err != nil {
		return err
	}
	var fi os.FileInfo
	fi, err = f.Stat()
	if err != nil {
		return err
	}
	//记录文件信息到当前对象
	r.file = f
	r.currentSize = fi.Size()
	r.todayEndUnix = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()).Unix()
	//检查是否开启日志缓冲
	if r.buffer == nil {
		r.w = r.file
		return nil
	}
	//重置缓冲区
	r.buffer.Reset(r.file)
	r.w = r.buffer
	return nil
}

func (r *File) Handle(record *contract.Record) bool {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()
	select {
	case <-r.closed:
		//强制返回false，让下一个日志handler继续处理日志信息
		return false
	default:
		break
	}
	//当前日志文件指针不存在，或者当前日志文件已经写满，或者日期已经跳到了第二天，则创建新的日志文件指针
	if r.file == nil || r.currentSize >= r.maxSize || record.Time.Unix() > r.todayEndUnix {
		var err error
		//关闭旧的日志文件
		err = r.closeFile()
		if err == nil {
			//寻找新的日志文件名
			var logName string
			logName, err = r.scanLogName(record.Time)
			if err == nil {
				//打开或创建日志文件
				err = r.openFile(logName, record.Time)
			}
		}
		if err != nil {
			libLog.Println(err)
			//强制返回false，让下一个日志handler继续处理日志信息
			return false
		}
	}
	//写入日志
	if n, err := r.formatter.ToWriter(r.w, record); err == nil {
		r.currentSize += n
		return r.bubble
	} else {
		//错误，调用标准库日志打印错误
		libLog.Println(err)
		//强制返回false，让下一个日志handler继续处理日志信息
		return false
	}
}