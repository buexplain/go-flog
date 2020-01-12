package formatter

import (
	"fmt"
	"github.com/buexplain/go-flog"
	"github.com/buexplain/go-flog/constant"
	"io"
	"strings"
	"time"
)

//行化日志结构体
type Line struct {
	eof        []byte
	timeFormat string
}

func NewLine() *Line {
	tmp := new(Line)
	tmp.eof = []byte(constant.EOF)
	tmp.timeFormat = time.RFC3339
	return tmp
}

func (this *Line) SetEOF(eof []byte) *Line {
	this.eof = eof
	return this
}

func (this *Line) SetTimeFormat(format string) *Line {
	this.timeFormat = format
	return this
}

func (this *Line) Format(w io.Writer, record *flog.Record) (n int, err error) {
	a := make([]interface{}, 0, 2)
	if record.Context != nil {
		a = append(a, record.Context)
	}
	if len(record.Extra) != 0 {
		a = append(a, record.Extra)
	}
	if len(a) == 0 {
		return fmt.Fprintf(w, "[%s] %s.%s: %s %s", record.Time.Format(this.timeFormat), record.Channel, record.LevelName, record.Message, this.eof)
	}
	f := "[%s] %s.%s: %s" + strings.Repeat(" %+v", len(a)) + "%s"
	b := make([]interface{}, 0, 7)
	b = append(b, record.Time.Format(this.timeFormat))
	b = append(b, record.Channel)
	b = append(b, record.LevelName)
	b = append(b, record.Message)
	b = append(b, a...)
	b = append(b, this.eof)
	return fmt.Fprintf(w, f, b...)
}
