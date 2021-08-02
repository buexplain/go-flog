package formatter

import (
	"bytes"
	"encoding/json"
	"github.com/buexplain/go-flog/contract"
	"io"
)

// JSON json化日志结构体
type JSON struct {
	prefix     string
	indent     string
	escapeHTML bool
}

func NewJSON() *JSON {
	tmp := &JSON{prefix: "", indent: "", escapeHTML: true}
	return tmp
}

func (r *JSON) SetIndent(prefix, indent string) *JSON {
	r.prefix = prefix
	r.indent = indent
	return r
}

func (r *JSON) SetEscapeHTML(on bool) *JSON {
	r.escapeHTML = on
	return r
}

func (r *JSON) ToBuffer(record *contract.Record) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	e := json.NewEncoder(buf)
	e.SetIndent(r.prefix, r.indent)
	e.SetEscapeHTML(r.escapeHTML)
	if err := e.Encode(record); err != nil {
		return nil, err
	}
	return buf, nil
}

func (r *JSON) ToWriter(w io.Writer, record *contract.Record) (written int64, err error) {
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
