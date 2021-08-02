package formatter

import (
	"bytes"
	"fmt"
	"github.com/buexplain/go-flog/contract"
	"io"
	"time"
)

// Line 行化日志结构体
type Line struct {
	eof        []byte
	timeFormat string
}

func NewLine() *Line {
	tmp := new(Line)
	tmp.eof = []byte{'\n'}
	tmp.timeFormat = time.RFC3339Nano
	return tmp
}

func (r *Line) SetTimeFormat(format string) *Line {
	r.timeFormat = format
	return r
}

func (r *Line) format(record *contract.Record) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	buf.WriteByte('[')
	buf.WriteString(record.Time.Format(r.timeFormat))
	buf.WriteByte(']')
	buf.WriteByte(' ')
	if len(record.Channel) > 0 {
		buf.WriteString(record.Channel)
		buf.WriteByte('.')
	}
	buf.WriteString(record.Level)
	buf.WriteByte(' ')
	buf.WriteString(record.Message)
	if record.Context != nil {
		_, _ = fmt.Fprintf(buf, " %+v", record.Context)
	}
	if record.Extra != nil {
		l := len(record.Extra)
		i := 1
		for k, v := range record.Extra {
			if i == l {
				_, _ = fmt.Fprintf(buf, " %s: %+v", k, v)
			}else {
				_, _ = fmt.Fprintf(buf, " %s: %+v,", k, v)
			}
			i++
		}
	}
	buf.WriteByte('\n')
	return
}

func (r *Line) ToBuffer(record *contract.Record) (buf *bytes.Buffer, err error) {
	return r.format(record)
}

func (r *Line) ToWriter(w io.Writer, record *contract.Record) (written int64, err error) {
	var buf *bytes.Buffer
	buf, err = r.format(record)
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
