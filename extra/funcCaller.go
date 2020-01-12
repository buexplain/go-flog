package extra

import (
	"github.com/buexplain/go-flog"
	"runtime"
)

//函数调用文件与行号
type FuncCaller struct {
	skip int
}

func NewFuncCaller(skip ...int) *FuncCaller {
	if len(skip) == 0 {
		skip = append(skip, 3)
	}
	return &FuncCaller{skip: skip[0]}
}

func (this *FuncCaller) SetSkip(skip int) *FuncCaller {
	this.skip = skip
	return this
}

func (this *FuncCaller) Processor(record *flog.Record) {
	if _, file, line, ok := runtime.Caller(this.skip); ok {
		record.Extra["File"] = file
		record.Extra["Line"] = line
	}
}
