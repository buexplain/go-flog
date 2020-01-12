package flog

import "io"

//日志格式化接口
type FormatterInterface interface {
	Format(w io.Writer, record *Record) (n int, err error)
}
