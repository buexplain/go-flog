package contract

import "time"

//日志信息结构体
type Record struct {
	//渠道
	Channel string
	//等级名称
	Level string
	//信息
	Message string
	//上下文
	Context interface{}
	//附加信息
	Extra map[string]interface{}
	//时间
	Time time.Time
}

func NewRecord() *Record {
	return &Record{Extra: make(map[string]interface{}), Time: time.Now()}
}
