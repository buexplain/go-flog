package flog

//添加日志额外信息的接口
type ExtraInterface interface {
	Processor(record *Record)
}
