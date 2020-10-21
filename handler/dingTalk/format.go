package dingtalk

import (
	"encoding/json"
	"fmt"
	"github.com/buexplain/go-flog"
	"io"
	"strings"
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
}

func NewFormatText() *FormatText {
	return &FormatText{
		At: &At{AtMobiles: []string{}, IsAtAll: false},
	}
}

func (r *FormatText) Format(w io.Writer, record *flog.Record) (n int, err error) {
	body := Text{}
	body.MsgType = "text"
	body.At = &(*r.At)
	body.Text = &Content{
		Content: "",
	}
	s := &strings.Builder{}
	_, _ = fmt.Fprintf(s, "[%s] %s: %s", record.Time.Format("2006-01-02 15:04:05"), record.LevelName, record.Message)
	if record.Context != nil {
		_, _ = fmt.Fprintf(s, "\n%+v", record.Context)
	}
	if record.Extra != nil && len(record.Extra) > 0 {
		for k, v := range record.Extra {
			_, _ = fmt.Fprintf(s, "\n%s: %+v", k, v)
		}
	}
	body.Text.Content = s.String()
	e := json.NewEncoder(w)
	return 0, e.Encode(body)
}
