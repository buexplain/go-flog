package dingtalk

import (
	"github.com/buexplain/go-flog"
)

type DingTalk struct {
	level     flog.Level
	robots    []*Robot
	index     int
	compress  map[string]byte
	timestamp int64
}

func New(level flog.Level, robots []*Robot, compress bool) *DingTalk {
	tmp := new(DingTalk)
	tmp.level = level
	tmp.robots = make([]*Robot, 0, len(robots))
	tmp.robots = append(tmp.robots, robots...)
	if compress {
		tmp.compress = map[string]byte{}
	} else {
		tmp.compress = nil
	}
	return tmp
}

//处理器入口
func (r *DingTalk) Handle(record *flog.Record) bool {
	if r.compress != nil {
		t := record.Time.Unix()
		if t-r.timestamp > 60 {
			r.compress = map[string]byte{}
			r.timestamp = t
		}
		if _, ok := r.compress[record.Message]; ok {
			return true
		}
		r.compress[record.Message] = 0
	}
	r.robots[r.index].send(record)
	r.index++
	if r.index == len(r.robots) {
		r.index = 0
	}
	return true
}

//判断当前处理器是否可以处理日志
func (r *DingTalk) IsHandling(level flog.Level) bool {
	return level <= r.level
}

//关闭日志处理器
func (r *DingTalk) Close() error {
	return nil
}
