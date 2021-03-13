package extra_test

import (
	"github.com/buexplain/go-flog/contract"
	"github.com/buexplain/go-flog/extra"
	"strings"
	"testing"
)

func TestFuncCaller(t *testing.T) {
	funcCaller := extra.NewFuncCaller()
	funcCaller.SetSkip(1)
	record := &contract.Record{Extra: map[string]interface{}{}}
	funcCaller.Processor(record)
	if line, ok := record.Extra["Line"]; !ok || line.(int) != 14 {
		t.Error("获取所属行号失败")
		return
	}
	if file, ok := record.Extra["File"]; !ok || !strings.HasSuffix(file.(string), "funcCaller_test.go") {
		t.Error("获取所属文件失败")
		return
	}
	t.Log(record.Extra["File"], record.Extra["Line"])
}
