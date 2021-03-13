package dingtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buexplain/go-flog/contract"
	"io"
	"strings"
	"time"
)

type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

type Content struct {
	Content string `json:"content"`
}

type Text struct {
	MsgType string `json:"msgtype"`
	At      *At
	Text    *Content `json:"text"`
}

type FormatText struct {
	At *At
	TimeFormat string
}

func (r *FormatText) SetAtMobile(mobile string) *FormatText {
	r.At.AtMobiles = append(r.At.AtMobiles, mobile)
	return r
}


func (r *FormatText) SetIsAtAll(isAtAll bool) *FormatText {
	r.At.IsAtAll = isAtAll
	return r
}

func NewFormatText() *FormatText {
	return &FormatText{
		At: &At{AtMobiles: []string{}, IsAtAll: false},
		TimeFormat: time.RFC3339Nano,
	}
}

func (r *FormatText) format(record *contract.Record) (buf *bytes.Buffer, err error) {
	body := Text{}
	body.MsgType = "text"
	body.At = &(*r.At)
	body.Text = &Content{
		Content: "",
	}
	s := &strings.Builder{}
	s.WriteByte('[')
	s.WriteString(record.Time.Format(r.TimeFormat))
	s.WriteByte(']')
	s.WriteByte(' ')
	s.WriteString(record.Channel)
	s.WriteByte('.')
	s.WriteString(record.Level)
	s.WriteByte(' ')
	s.WriteString(record.Message)
	if record.Context != nil {
		_, _ = fmt.Fprintf(s, "\n%+v", record.Context)
	}
	if record.Extra != nil && len(record.Extra) > 0 {
		for k, v := range record.Extra {
			_, _ = fmt.Fprintf(s, "\n%s: %+v", k, v)
		}
	}
	body.Text.Content = s.String()
	buf = bytes.NewBuffer(nil)
	e := json.NewEncoder(buf)
	err = e.Encode(body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (r *FormatText) ToBuffer(record *contract.Record) (buf *bytes.Buffer, err error) {
	return r.format(record)
}

func (r *FormatText) ToWriter(w io.Writer, record *contract.Record) (written int64, err error) {
	var buf *bytes.Buffer
	buf, err = r.ToBuffer(record)
	if err != nil {
		return 0, err
	}
	var n int
	n , err = w.Write(buf.Bytes())
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}
