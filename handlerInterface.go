package flog

//日志处理器接口
type HandlerInterface interface {
	//处理器入口
	Handle(record *Record) bool
	//判断当前处理器是否可以处理日志
	IsHandling(level Level) bool
	//关闭日志处理器
	Close() error
}
