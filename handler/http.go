package handler

import (
	"bytes"
	"github.com/buexplain/go-flog/contract"
	formatter2 "github.com/buexplain/go-flog/formatter"
	"io/ioutil"
	libLog "log"
	"net/http"
	"net/url"
	"time"
)

// HTTP http接口日志处理器
type HTTP struct {
	//日志等级
	level contract.Level
	//日志格式化处理器
	formatter contract.Formatter
	//是否阻止进入下一个日志处理器
	bubble bool
	//日志写入地址
	url string
	//请求头部
	header http.Header
	//超时设置
	timeout time.Duration
}

func NewHTTP(level contract.Level, formatter contract.Formatter, url string) *HTTP {
	tmp := new(HTTP)
	tmp.level = level
	tmp.formatter = formatter
	tmp.bubble = false
	tmp.url = url
	tmp.header = make(http.Header, 0)
	if _, ok := formatter.(*formatter2.JSON); ok {
		tmp.header.Set("Content-Type", "application/json; charset=utf-8")
	}else {
		tmp.header.Set("Content-Type", "text/plain; charset=utf-8")
	}
	tmp.timeout = 5 * time.Second
	return tmp
}

func (r *HTTP) SetBubble(bubble bool) *HTTP {
	r.bubble = bubble
	return r
}

func (r *HTTP) SetHeader(h http.Header) *HTTP {
	r.header = h
	return r
}

func (r *HTTP) SetTimeout(t time.Duration) *HTTP {
	r.timeout = t
	return r
}

func (r *HTTP) Close() error {
	return nil
}

// IsHandling 判断当前处理器是否可以处理日志
func (r *HTTP) IsHandling(level contract.Level) bool {
	return level <= r.level
}

// Handle 处理器入口
func (r *HTTP) Handle(record *contract.Record) bool {
	request, err := http.NewRequest(http.MethodPost, r.url, nil)
	if err != nil {
		libLog.Println(err)
		return false
	}

	//克隆头部信息
	for k, vv := range r.header {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		request.Header[k] = vv2
	}

	//设置请求体
	var buf *bytes.Buffer
	buf, err = r.formatter.ToBuffer(record)
	if err != nil {
		libLog.Println(err)
		return false
	}
	request.Body = ioutil.NopCloser(buf)

	client := http.Client{Timeout: r.timeout}
	_, err = client.Do(request)
	if err != nil {
		if e, ok := err.(*url.Error); !ok || !e.Timeout() {
			libLog.Println(err)
		}
		return false
	}

	return r.bubble
}
