package dingtalk_test

import (
	"context"
	"github.com/buexplain/go-flog/contract"
	dingTalk "github.com/buexplain/go-flog/handler/dingTalk"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestDingTalk(t *testing.T) {
	robots := make([]*dingTalk.Robot, 0)
	robots = append(robots, dingTalk.NewRobot("http://127.0.0.1:8107/test?token=aa", "one", dingTalk.NewFormatText().SetAtMobile("123456"), 3))
	robots = append(robots, dingTalk.NewRobot("http://127.0.0.1:8107/test?token=bb", "two", dingTalk.NewFormatText().SetIsAtAll(true), 3))
	dTalk := dingTalk.New(contract.LevelDebug, robots, false)
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		b, _ := ioutil.ReadAll(request.Body)
		t.Logf(string(b))
	})
	server := &http.Server{Addr: "127.0.0.1:8107", Handler: mux}
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
				dTalk.Handle(tmp)
			}
		}()
	}
	//等待写入完成
	wg.Wait()
	<-time.After(10*time.Second)
	for i:=0; i<10; i++ {
		go func() {
			_ = dTalk.Close()
		}()
	}
	<-time.After(1*time.Second)
	//停止http服务器
	_ = server.Shutdown(context.Background())
}