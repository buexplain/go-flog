package formatter_test

import (
	"bytes"
	"github.com/buexplain/go-flog/contract"
	"github.com/buexplain/go-flog/formatter"
	"testing"
)

func TestLine(t *testing.T) {
	record := contract.NewRecord()
	record.Extra["extraA"] = "extra"
	record.Extra["extraB"] = 100
	record.Extra["extraC"] = struct {
		Name string
		Age uint8
	}{
		Name: "西门吹雪",
		Age: 108,
	}
	//record.Channel = "channel"
	record.Message = "message"
	record.Context = struct {
		A string
		B struct{
			C int
		}
	}{A: "context", B: struct {
		C int
	}{C: 100}}
	record.Level = contract.GetNameByLevel(contract.LevelDebug)
	j := formatter.NewLine()
	buf := &bytes.Buffer{}
	i, err := j.ToWriter(buf, record)
	if err != nil {
		t.Error("line格式化失败：", err.Error())
		return
	}
	if i != int64(buf.Len()) {
		t.Error("line格式化的字符长度与返回的长度不一致")
		return
	}
	t.Log(i, buf.String())
}
