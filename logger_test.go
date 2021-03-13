package flog_test

import (
	"fmt"
	"github.com/buexplain/go-flog"
	"github.com/buexplain/go-flog/contract"
	"github.com/buexplain/go-flog/extra"
	"github.com/buexplain/go-flog/formatter"
	"github.com/buexplain/go-flog/handler"
	"os"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	path := "./test283782767"
	//异步的文件处理器
	fileAsync := handler.NewFile(contract.LevelDebug, formatter.NewLine(), path+"loggerAwaitFileAsync")
	fileAsync.SetMaxSize(11 << 20)
	fileAsync.SetBuffer(2 << 20, 3*time.Second)
	//同步的文件处理器
	fileAwait := handler.NewFile(contract.LevelDebug, formatter.NewLine(), path+"loggerAwaitFileAwait")
	fileAwait.SetMaxSize(11 << 20)
	//同步的日志组件
	loggerAwait := flog.New("await", nil, extra.NewFuncCaller(3))
	loggerAwait.PushHandler(fileAsync)
	loggerAwait.PushHandler(fileAwait)
	//异步的文件处理器
	fileAsync = handler.NewFile(contract.LevelDebug, formatter.NewLine(), path+"loggerAsyncFileAsync")
	fileAsync.SetMaxSize(11 << 20)
	fileAsync.SetBuffer(2 << 20, 3*time.Second)
	//同步的文件处理器
	fileAwait = handler.NewFile(contract.LevelDebug, formatter.NewLine(), path+"loggerAsyncFileAwait")
	fileAwait.SetMaxSize(11 << 20)
	//异步的日志组件
	loggerAsync := flog.New("async", nil, extra.NewFuncCaller(3))
	loggerAsync.PushHandler(fileAsync)
	loggerAsync.PushHandler(fileAwait)
	loggerAsync.Async(10000)
	type Info struct {
		Name string
		Age int
	}
	wg := &sync.WaitGroup{}
	for i:=0; i<10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for i:=0; i<1000000; i++ {
				loggerAsync.Debug(fmt.Sprintf("go %d message %d", index, i), Info{Name: "刘备", Age: 28})
				loggerAsync.Error(fmt.Sprintf("go %d message %d", index, i), Info{Name: "关羽", Age: 28})
				loggerAsync.AlertF("go %d message %d %+v", index, i, Info{Name: "张飞", Age: 28})
			}
		}(i)
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for i:=0; i<1000000; i++ {
				loggerAwait.Debug(fmt.Sprintf("go %d message %d", index, i), Info{Name: "刘备", Age: 28})
				loggerAwait.Error(fmt.Sprintf("go %d message %d", index, i), Info{Name: "关羽", Age: 28})
				loggerAwait.AlertF("go %d message %d %+v", index, i, Info{Name: "张飞", Age: 28})
			}
		}(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-time.After(time.Second*10)
			if err := loggerAsync.Close(); err != nil {
				t.Log("关闭日志组件loggerAsync失败", err)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-time.After(time.Second*10)
			if err := loggerAwait.Close(); err != nil {
				t.Log("关闭日志组件loggerAwait失败", err)
			}
		}()
	}
	wg.Wait()
	if err := os.RemoveAll(path+"loggerAwaitFileAsync"); err != nil {
		t.Error("日志处理器的文件未关闭 loggerAwaitFileAsync：", err)
	}
	if err := os.RemoveAll(path+"loggerAwaitFileAwait"); err != nil {
		t.Error("日志处理器的文件未关闭 loggerAwaitFileAwait：", err)
	}
	if err := os.RemoveAll(path+"loggerAsyncFileAsync"); err != nil {
		t.Error("日志处理器的文件未关闭 loggerAsyncFileAsync：", err)
	}
	if err := os.RemoveAll(path+"loggerAsyncFileAwait"); err != nil {
		t.Error("日志处理器的文件未关闭 loggerAsyncFileAwait：", err)
	}
}