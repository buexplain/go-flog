package handler

import (
	"github.com/buexplain/go-flog/contract"
	"os"
)

//标准输出与标准出错日志处理器
type STD struct {
	//日志等级
	level contract.Level
	//日志格式化处理器
	formatter contract.Formatter
	//是否阻止进入下一个日志处理器
	bubble bool
	//标准输出与标准错误分割的日志等级
	dst contract.Level
}

func NewSTD(level contract.Level, formatter contract.Formatter, dst contract.Level) *STD {
	tmp := new(STD)
	tmp.level = level
	tmp.formatter = formatter
	tmp.bubble = false
	tmp.dst = dst
	return tmp
}

func (r *STD) SetBubble(bubble bool) *STD {
	r.bubble = bubble
	return r
}

func (r *STD) Close() error {
	return nil
}

//判断当前处理器是否可以处理日志
func (r *STD) IsHandling(level contract.Level) bool {
	return level <= r.level
}

//处理器入口
func (r *STD) Handle(record *contract.Record) bool {
	var err error
	if r.dst == -1 {
		_, err = r.formatter.ToWriter(os.Stdout, record)
	} else {
		if contract.GetLevelByName(record.Level) <= r.dst {
			_, err = r.formatter.ToWriter(os.Stderr, record)
		} else {
			_, err = r.formatter.ToWriter(os.Stdout, record)
		}
	}
	if err != nil {
		//强制返回false
		//让下一个日志handler继续处理日志信息
		return false
	}
	return r.bubble
}
