package formatter

import (
	"encoding/json"
	"github.com/buexplain/go-flog"
	"io"
)

//json化日志结构体
type JSON struct {
	prefix     string
	indent     string
	escapeHTML bool
}

func NewJSON() *JSON {
	tmp := &JSON{prefix: "", indent: "", escapeHTML: true}
	return tmp
}

func (this *JSON) SetIndent(prefix, indent string) *JSON {
	this.prefix = prefix
	this.indent = indent
	return this
}

func (this *JSON) SetEscapeHTML(on bool) *JSON {
	this.escapeHTML = on
	return this
}

func (this *JSON) Format(w io.Writer, record *flog.Record) (int, error) {
	e := json.NewEncoder(w)
	e.SetIndent(this.prefix, this.indent)
	e.SetEscapeHTML(this.escapeHTML)
	return 0, e.Encode(record)
}
