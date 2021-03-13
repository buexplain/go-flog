package handler_test

import (
	"github.com/buexplain/go-flog/contract"
	"github.com/buexplain/go-flog/formatter"
	"github.com/buexplain/go-flog/handler"
	"sync"
	"testing"
)

func TestSTD(t *testing.T) {
	std := handler.NewSTD(contract.LevelDebug, formatter.NewLine(), contract.LevelError)
	if !std.IsHandling(contract.LevelEmergency) || !std.IsHandling(contract.LevelDebug) {
		t.Error("std输出的日志等级校验失败")
		return
	}
	wg := &sync.WaitGroup{}
	//并发写入
	for i:=0; i<10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			record := contract.NewRecord()
			record.Extra["extraA"] = "extra"
			record.Extra["extraB"] = 100
			record.Extra["extraC"] = struct {
				Name string
				Age uint8
			}{
				Name: "西门吹雪",
				Age: 108,
			}
			record.Channel = "channel"
			record.Message = "message"
			record.Context = struct {
				A string
				B struct{
					C int
				}
			}{A: "context", B: struct {
				C int
			}{C: 100}}
			level := []contract.Level{contract.LevelDebug, contract.LevelInfo, contract.LevelError, contract.LevelAlert}
			for _, v := range level {
				tmp := &(*record)
				tmp.Level = contract.GetNameByLevel(v)
				std.Handle(tmp)
			}
		}()
	}
	wg.Wait()
}


