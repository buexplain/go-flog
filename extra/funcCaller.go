package extra

import (
	"github.com/buexplain/go-flog/contract"
	"runtime"
)

// FuncCaller 函数调用文件与行号
type FuncCaller struct {
	skip int
}

func NewFuncCaller(skip ...int) *FuncCaller {
	if len(skip) == 0 {
		skip = append(skip, 3)
	}
	return &FuncCaller{skip: skip[0]}
}

func (r *FuncCaller) SetSkip(skip int) *FuncCaller {
	r.skip = skip
	return r
}

func (r *FuncCaller) Processor(record *contract.Record) {
	if _, file, line, ok := runtime.Caller(r.skip); ok {
		record.Extra["File"] = file
		record.Extra["Line"] = line
	}
}
