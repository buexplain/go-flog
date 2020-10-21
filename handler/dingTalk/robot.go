package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/buexplain/go-flog"
	"io/ioutil"
	libLog "log"
	"net/http"
	"strconv"
	"time"
)

type Robot struct {
	url       string
	secret    []byte
	formatter flog.FormatterInterface
	recordCh  chan *flog.Record
}

func NewRobot(url string, secret string, formatter flog.FormatterInterface) *Robot {
	tmp := &Robot{}
	tmp.url = url
	if secret == "" {
		tmp.secret = nil
	} else {
		tmp.secret = []byte(secret)
	}
	tmp.recordCh = make(chan *flog.Record, 200)
	tmp.formatter = formatter
	tmp.gof()
	return tmp
}

func (r *Robot) gof() {
	go func() {
		defer func() {
			if re := recover(); re != nil {
				//如果异常退出，则间隔一段时间后重写启动一条协程
				<-time.After(10 * time.Second)
				r.gof()
			}
		}()
		//钉钉群机器人限制频率为一分钟20条，此处每三秒发送一次消息
		tick := time.Tick(3 * time.Second)
		for _ = range tick {
			for record := range r.recordCh {
				buf := &bytes.Buffer{}
				if _, err := r.formatter.Format(buf, record); err != nil {
					libLog.Println(err)
					break
				}
				req, err := http.NewRequest(http.MethodPost, r.makeURL(), nil)
				if err != nil {
					libLog.Println(err)
					break
				}
				req.Body = ioutil.NopCloser(buf)
				req.Header.Add("Content-Type", "application/json;charset=utf-8")
				client := &http.Client{Timeout: time.Second * 3}
				if _, err := client.Do(req); err != nil {
					libLog.Println(err)
				}
				break
			}
		}
	}()
}

func (r *Robot) send(record *flog.Record) bool {
	select {
	case r.recordCh <- record:
		return true
	default:
		return false
	}
}

func (r *Robot) makeURL() string {
	url := r.url
	if r.secret != nil && len(r.secret) > 0 {
		timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
		buf := bytes.NewBuffer(nil)
		buf.WriteString(timestamp)
		buf.WriteString("\n")
		buf.Write(r.secret)
		h := hmac.New(sha256.New, r.secret)
		h.Write(buf.Bytes())
		sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
		url = fmt.Sprintf("%s&timestamp=%s&sign=%s", url, timestamp, sign)
	}
	return url
}
