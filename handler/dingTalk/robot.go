package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/buexplain/go-flog/contract"
	"io/ioutil"
	libLog "log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Robot struct {
	url       string
	secret    []byte
	formatter contract.Formatter
	recordCh  chan contract.Record
	closed chan struct{}
}

func NewRobot(url string, secret string, formatter contract.Formatter, capacity int) *Robot {
	tmp := &Robot{}
	tmp.url = url
	if secret == "" {
		tmp.secret = nil
	} else {
		tmp.secret = []byte(secret)
	}
	tmp.formatter = formatter
	tmp.recordCh = make(chan contract.Record, capacity)
	tmp.closed = make(chan struct{})
	tmp.gof()
	return tmp
}

func (r *Robot) close() {
	close(r.closed)
}

func (r *Robot) gof() {
	go func() {
		//钉钉群机器人限制频率为一分钟20条，此处每三秒发送一次消息
		tick := time.NewTicker(3 * time.Second)
		defer tick.Stop()
		defer func() {
			if re := recover(); re != nil {
				//如果异常退出，则间隔一段时间后重启动一条协程
				<-time.After(10 * time.Second)
				r.gof()
			}else {
				close(r.recordCh)
			}
		}()
		var buf *bytes.Buffer
		var err error
		var req *http.Request
		for {
			buf = nil
			err = nil
			req = nil
			select {
			case <-r.closed:
				return
			case <-tick.C:
				select {
				case <-r.closed:
					return
				case record := <- r.recordCh:
					buf, err = r.formatter.ToBuffer(&record)
					if err != nil {
						libLog.Println(err)
						break
					}
					req, err = http.NewRequest(http.MethodPost, r.makeURL(), ioutil.NopCloser(buf))
					if err != nil {
						libLog.Println(err)
						break
					}
					req.Header.Add("Content-Type", "application/json;charset=utf-8")
					client := http.Client{Timeout: time.Second * 5}
					if _, err := client.Do(req); err != nil {
						if e, ok := err.(*url.Error); !ok || !e.Timeout() {
							libLog.Println(err)
						}
					}
				}
			}
		}
	}()
}

func (r *Robot) send(record *contract.Record) bool {
	select {
	case <-r.closed:
		return false
	default:
		select {
		case r.recordCh <- *record:
			return true
		default:
			return false
		}
	}
}

func (r *Robot) makeURL() string {
	if r.secret == nil || len(r.secret) == 0 {
		return r.url
	}
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	buf := bytes.NewBuffer(nil)
	buf.WriteString(timestamp)
	buf.WriteByte('\n')
	buf.Write(r.secret)
	h := hmac.New(sha256.New, r.secret)
	h.Write(buf.Bytes())
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s&timestamp=%s&sign=%s", r.url, timestamp, sign)
}