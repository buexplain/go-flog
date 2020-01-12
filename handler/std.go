package handler

import (
	"github.com/buexplain/go-flog"
	libLog "log"
	"os"
)

//标准输出与标准出错日志处理器
type STD struct {
	//日志等级
	level flog.Level
	//日志格式化处理器
	formatter flog.FormatterInterface
	//是否阻止进入下一个日志处理器
	bubble bool
	//标准输出与标准错误分割的日志等级
	splitLevel flog.Level
}

func NewSTD(level flog.Level, formatter flog.FormatterInterface, splitLevel flog.Level) *STD {
	tmp := new(STD)
	tmp.level = level
	tmp.formatter = formatter
	tmp.bubble = false
	tmp.splitLevel = splitLevel
	return tmp
}

func (this *STD) SetBubble(bubble bool) *STD {
	this.bubble = bubble
	return this
}

func (this *STD) Close() error {
	return nil
}

//判断当前处理器是否可以处理日志
func (this *STD) IsHandling(level flog.Level) bool {
	return level <= this.level
}

//处理器入口
func (this *STD) Handle(record *flog.Record) bool {
	var err error
	if this.splitLevel < 0 {
		//全部都打印到标准输出
		_, err = this.formatter.Format(os.Stdout, record)
	} else {
		if record.Level <= this.splitLevel {
			//打印到标准出错
			_, err = this.formatter.Format(os.Stderr, record)
		} else {
			//打印到标准输出
			_, err = this.formatter.Format(os.Stdout, record)
		}
	}
	if err != nil {
		//错误，调用标准库日志打印错误
		libLog.Println(err)
		//强制返回false，让下一个日志handler继续处理日志信息
		return false
	}
	return this.bubble
}
