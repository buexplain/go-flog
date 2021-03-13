package handler_test

import (
	"context"
	"github.com/buexplain/go-flog/contract"
	"github.com/buexplain/go-flog/formatter"
	"github.com/buexplain/go-flog/handler"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestHTTP(t *testing.T) {
	std := handler.NewHTTP(contract.LevelDebug, formatter.NewJSON(), "http://127.0.0.1:8106/test")
	if !std.IsHandling(contract.LevelEmergency) || !std.IsHandling(contract.LevelDebug) {
		t.Error("http输出的日志等级校验失败")
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		b, _ := ioutil.ReadAll(request.Body)
		t.Logf(string(b))
	})
	server := &http.Server{Addr: "127.0.0.1:8106", Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Error("启动http服务失败", err)
		}
	}()
	//等待http服务启动
	<- time.After(time.Second*1)
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
	//等待写入完成
	wg.Wait()
	//停止http服务器
	_ = server.Shutdown(context.Background())
}


