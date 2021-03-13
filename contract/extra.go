package contract

//添加日志额外信息的接口
type Extra interface {
	Processor(record *Record)
}
