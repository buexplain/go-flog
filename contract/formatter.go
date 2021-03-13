package contract

import (
	"bytes"
	"io"
)

//日志格式化接口
type Formatter interface {
	ToWriter(w io.Writer, record *Record) (written int64, err error)
    ToBuffer(record *Record) (buf *bytes.Buffer, err error)
}
