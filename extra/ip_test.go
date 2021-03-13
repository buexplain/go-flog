package extra_test

import (
	"github.com/buexplain/go-flog/contract"
	"github.com/buexplain/go-flog/extra"
	"testing"
)

func TestIP(t *testing.T) {
	ip := extra.NewIP()
	if ip == "" {
		t.Error("获取ip失败")
		return
	}
	record := &contract.Record{Extra: map[string]interface{}{}}
	ip.Processor(record)
	if record.Extra["IP"].(string) != string(ip) {
		t.Error("设置ip失败")
		return
	}
	t.Log("ip: ", ip)
}