package handler

import (
	"bytes"
	"github.com/buexplain/go-flog"
	"io/ioutil"
	libLog "log"
	"net/http"
	"time"
)

//http接口日志处理器
type HTTP struct {
	//日志等级
	level flog.Level
	//日志格式化处理器
	formatter flog.FormatterInterface
	//是否阻止进入下一个日志处理器
	bubble bool
	//日志写入地址
	url string
	//请求头部
	header http.Header
	//超时设置
	timeout time.Duration
}

func NewHTTP(level flog.Level, formatter flog.FormatterInterface, url string) *HTTP {
	tmp := new(HTTP)
	tmp.level = level
	tmp.formatter = formatter
	tmp.bubble = false
	tmp.url = url
	tmp.header = make(http.Header, 0)
	tmp.header.Set("Content-Type", "text/plain; charset=utf-8")
	tmp.timeout = time.Duration(1500 * time.Millisecond)
	return tmp
}

func (this *HTTP) SetBubble(bubble bool) *HTTP {
	this.bubble = bubble
	return this
}

func (this *HTTP) SetHeader(h http.Header) *HTTP {
	this.header = h
	return this
}

func (this *HTTP) SetTimeout(t time.Duration) *HTTP {
	this.timeout = t
	return this
}

func (this *HTTP) Close() error {
	return nil
}

//判断当前处理器是否可以处理日志
func (this *HTTP) IsHandling(level flog.Level) bool {
	return level <= this.level
}

//处理器入口
func (this *HTTP) Handle(record *flog.Record) bool {
	request, err := http.NewRequest(http.MethodPost, this.url, nil)
	if err != nil {
		//错误，调用标准库日志打印错误
		libLog.Println(err)
		//强制返回false，让下一个日志handler继续处理日志信息
		return false
	}

	//克隆头部信息
	for k, vv := range this.header {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		request.Header[k] = vv2
	}

	//设置请求体
	buf := &bytes.Buffer{}
	if _, err := this.formatter.Format(buf, record); err != nil {
		//错误，调用标准库日志打印错误
		libLog.Println(err)
		//强制返回false，让下一个日志handler继续处理日志信息
		return false
	}
	request.Body = ioutil.NopCloser(buf)

	client := &http.Client{Timeout: this.timeout}
	_, err = client.Do(request)
	if err != nil {
		//错误，调用标准库日志打印错误
		libLog.Println(err)
		//强制返回false，让下一个日志handler继续处理日志信息
		return false
	}

	return this.bubble
}
