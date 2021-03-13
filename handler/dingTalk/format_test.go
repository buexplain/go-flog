package dingtalk_test

import (
	"bytes"
	"github.com/buexplain/go-flog/contract"
	dingtalk "github.com/buexplain/go-flog/handler/dingTalk"
	"sync"
	"testing"
)

func TestFormatText(t *testing.T) {
	formatText := dingtalk.NewFormatText()
	formatText.At.IsAtAll = true
	formatText.At.AtMobiles = append(formatText.At.AtMobiles, "123456")
	wg := &sync.WaitGroup{}
	for i:=0; i<10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
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
			record.Channel = "channel"
			record.Message = "message"
			record.Context = struct {
				A string
				B struct{
					C int
				}
			}{A: "context", B: struct {
				C int
			}{C: 100}}
			level := []contract.Level{contract.LevelDebug, contract.LevelInfo, contract.LevelError, contract.LevelAlert}
			var err error
			var buf *bytes.Buffer
			for _, v := range level {
				tmp := &(*record)
				tmp.Level = contract.GetNameByLevel(v)
				buf, err = formatText.ToBuffer(record)
				if err != nil {
					t.Error("钉钉text格式化失败", err)
				}
			}
			if index == 0 {
				t.Log(buf.String())
			}
		}(i)
	}
	wg.Wait()
}
