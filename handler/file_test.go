package handler_test

import (
	"github.com/buexplain/go-flog/contract"
	"github.com/buexplain/go-flog/formatter"
	"github.com/buexplain/go-flog/handler"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

//测试同步写入日志
func TestFileAwait(t *testing.T) {
	path, err := ioutil.TempDir("./", "test")
	if err != nil {
		t.Error("构建临时目录失败")
	}
	file := handler.NewFile(contract.LevelDebug, formatter.NewLine(), path)
	file.SetMaxSize(1024*1024*256)
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
		file.Handle(tmp)
	}
	//此时文件有数据，检查是否符合要求
	var fi os.FileInfo
	if m, err := filepath.Glob(filepath.Join(path, "*.log")); err != nil {
		t.Error("获取日志处理的结果失败：", err)
	}else if len(m) != 1 {
		t.Error("获取日志处理的输出结果不一致：", err)
	}else {
		fi, err = os.Stat(m[0])
		if err != nil {
			t.Error("获取日志处理的输出结果失败", err)
			return
		}
		if fi.Size() == 0 {
			t.Error("日志处理的输出结果错误")
			return
		}
	}
	//关闭日志处理器，此时并不会改变现有日志文件的大小，因为没有异步的情况
	if err  := file.Close(); err != nil {
		t.Error("日志处理测试失败：", err)
	}
	//检查磁盘的日志文件大小
	oldSize := fi.Size()
	fi, err = os.Stat(filepath.Join(path, fi.Name()))
	if err != nil {
		t.Error("获取日志处理的输出结果失败", err)
		return
	}
	if fi.Size() != oldSize {
		t.Error("日志处理的输出结果错误")
		return
	}
	if err := os.RemoveAll(path); err != nil {
		t.Error("日志处理器的文件未关闭：", err)
	}
}

//测试异步缓冲关闭时候落盘
func TestFileAsyncClose(t *testing.T) {
	path, err := ioutil.TempDir("./", "test")
	if err != nil {
		t.Error("构建临时目录失败")
	}
	file := handler.NewFile(contract.LevelDebug, formatter.NewJSON(), path)
	//设置一个2MiB的缓冲器，30秒刷新一次，保证日志都会写入到缓冲器中
	file.SetBuffer(1024*1024*2, time.Second*30)
	//设置一个超大的文件大小限制值，保证日志都冲刷到一个日志文件中，而不是触发文件大小限制导致切换文件的过程中强制刷新
	file.SetMaxSize(1024*1024*256)
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
		file.Handle(tmp)
	}
	<- time.After(time.Second*1)
	//此时文件没有数据，检查是否符合要求
	var fi os.FileInfo
	if m, err := filepath.Glob(filepath.Join(path, "*.log")); err != nil {
		t.Error("获取日志处理的结果失败：", err)
	}else if len(m) != 1 {
		t.Error("获取日志处理的输出结果不一致：", err)
	}else {
		fi, err = os.Stat(m[0])
		if err != nil {
			t.Error("获取日志处理的输出结果失败", err)
			return
		}
		if fi.Size() > 0 {
			t.Error("日志处理的输出结果错误")
			return
		}
	}
	//关闭日志处理器，此时会强制刷新到磁盘
	if err  := file.Close(); err != nil {
		t.Error("日志处理测试失败：", err)
	}
	//检查磁盘的日志文件是否已经有内容
	fi, err = os.Stat(filepath.Join(path, fi.Name()))
	if err != nil {
		t.Error("获取日志处理的输出结果失败", err)
		return
	}
	if fi.Size() == 0 {
		t.Error("日志处理的输出结果错误")
		return
	}
	if err := os.RemoveAll(path); err != nil {
		t.Error("日志处理器的文件未关闭：", err)
	}
}

//测试异步缓冲定时刷新落盘
func TestFileAsyncTicker(t *testing.T) {
	path, err := ioutil.TempDir("./", "test")
	if err != nil {
		t.Error("构建临时目录失败")
	}
	file := handler.NewFile(contract.LevelDebug, formatter.NewJSON(), path)
	//设置一个2MiB的缓冲器，每三秒定时刷新一次
	file.SetBuffer(1024*1024*2, time.Second*3)
	//设置一个超大的文件大小限制值，保证日志都冲刷到一个日志文件中，而不是触发文件大小限制导致切换文件的过程中强制刷新
	file.SetMaxSize(1024*1024*256)
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
		file.Handle(tmp)
	}
	//等待四秒，让定时器定时刷新数据到磁盘
	<- time.After(time.Second*4)
	if m, err := filepath.Glob(filepath.Join(path, "*.log")); err != nil {
		t.Error("获取日志处理的结果失败：", err)
	}else if len(m) != 1 {
		t.Error("获取日志处理的输出结果不一致：", err)
	}
	if err  := file.Close(); err != nil {
		t.Error("日志处理测试失败：", err)
	}
	if err := os.RemoveAll(path); err != nil {
		t.Error("日志处理器的文件未关闭：", err)
	}
}

//测试异步缓冲日志按限制大小切割，强制落盘
func TestFileAsyncSpilt(t *testing.T) {
	path, err := ioutil.TempDir("./", "test")
	if err != nil {
		t.Error("构建临时目录失败")
	}
	file := handler.NewFile(contract.LevelDebug, formatter.NewLine(), path)
	//设置一个极小字节、定时极长的缓冲器，保证每一条日志的写入都会撑爆缓冲器，迫使缓冲器主动刷新
	file.SetBuffer(10, time.Second*20)
	//设置一个极小的文件大小限制值，保证每一条日志都会写入到新的文件
	file.SetMaxSize(30)
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
		file.Handle(tmp)
	}
	//日志关闭之前检查输出结果，同时因为定时刷新的时间过长，所以这些输出文件
	if m, err := filepath.Glob(filepath.Join(path, "*.log")); err != nil {
		t.Error("获取日志处理的结果失败：", err)
	}else if len(m) != 4 {
		t.Error("获取日志处理的输出结果不一致：", err)
	}
	if err  := file.Close(); err != nil {
		t.Error("日志处理测试失败：", err)
	}
	if err := os.RemoveAll(path); err != nil {
		t.Error("日志处理器的文件未关闭：", err)
	}
}

//测试异步并发写入与关闭
func TestFileAsyncConcurrent(t *testing.T) {
	path, err := ioutil.TempDir("./", "test")
	if err != nil {
		t.Error("构建临时目录失败")
	}
	file := handler.NewFile(contract.LevelDebug, formatter.NewLine(), path)
	//设置强制冒泡，当异步冲刷协程收到close信号后会返回false
	file.SetBubble(true)
	//设置一个2MiB的缓冲器，每三秒定时刷新一次
	file.SetBuffer(1024*1024*2, time.Second*3)
	//设置一个适中的文件大小限制值，保证日志有切割的机会
	file.SetMaxSize(1024*1024*10)
	wg := &sync.WaitGroup{}
	//并发写入日志
	for i:=0 ; i<10; i++ {
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
			for i:=0; i<100000000; i++ {
				for _, v := range level {
					tmp := &(*record)
					tmp.Level = contract.GetNameByLevel(v)
					if !file.Handle(tmp) {
						//异步冲刷协程收到close信号后会返回false，停止写入日志
						return
					}
				}
			}
		}()
	}
	//并发关闭日志处理器
	for i:=0 ; i<10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-time.After(10*time.Second)
			if err := file.Close(); err != nil {
				t.Error("关闭日志处理器失败", err)
			}
		}()
	}
	wg.Wait()
	if err := os.RemoveAll(path); err != nil {
		t.Error("日志处理器的文件未关闭：", err)
	}
}

//测试同步并发写入与关闭
func TestFileAwaitConcurrent(t *testing.T) {
	path, err := ioutil.TempDir("./", "test")
	if err != nil {
		t.Error("构建临时目录失败")
	}
	file := handler.NewFile(contract.LevelDebug, formatter.NewLine(), path)
	//设置强制冒泡，当异步冲刷协程收到close信号后会返回false
	file.SetBubble(true)
	//设置一个适中的文件大小限制值，保证日志有切割的机会
	file.SetMaxSize(1024*1024*10)
	wg := &sync.WaitGroup{}
	//并发写入日志
	for i:=0 ; i<10; i++ {
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
			for i:=0; i<100000000; i++ {
				for _, v := range level {
					tmp := &(*record)
					tmp.Level = contract.GetNameByLevel(v)
					if !file.Handle(tmp) {
						//异步冲刷协程收到close信号后会返回false，停止写入日志
						return
					}
				}
			}
		}()
	}
	//并发关闭日志处理器
	for i:=0 ; i<10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-time.After(10*time.Second)
			if err := file.Close(); err != nil {
				t.Error("关闭日志处理器失败", err)
			}
		}()
	}
	wg.Wait()
	if err := os.RemoveAll(path); err != nil {
		t.Error("日志处理器的文件未关闭：", err)
	}
}