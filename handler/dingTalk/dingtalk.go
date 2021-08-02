package dingtalk

import (
	"github.com/buexplain/go-flog/contract"
	"sync"
)

type DingTalk struct {
	level     contract.Level
	robotCh   chan *Robot
	robots    []*Robot
	writeLock *sync.Mutex
	compress  bool
	compressed map[string]byte
	timestamp int64
}

func New(level contract.Level, robots []*Robot, compress bool) *DingTalk {
	tmp := new(DingTalk)
	tmp.level = level
	tmp.robotCh = make(chan *Robot, len(robots))
	tmp.robots = make([]*Robot, 0, len(robots))
	tmp.writeLock = new(sync.Mutex)
	for _, robot := range robots {
		tmp.robotCh <- robot
		tmp.robots = append(tmp.robots, robot)
	}
	tmp.compress = compress
	if tmp.compress {
		tmp.compressed = map[string]byte{}
	} else {
		tmp.compressed = nil
	}
	return tmp
}

// Handle 处理器入口
func (r *DingTalk) Handle(record *contract.Record) bool {
	if r.compress {
		r.writeLock.Lock()
		defer r.writeLock.Unlock()
		t := record.Time.Unix()
		if t - r.timestamp > 60 || len(r.compressed) > 10000 {
			r.compressed = map[string]byte{}
			r.timestamp = t
		}else {
			if _, ok := r.compressed[record.Message]; ok {
				//强制进入下一个日志处理器
				return true
			}
		}
		r.compressed[record.Message] = 0
	}
	robot := <-r.robotCh
	r.robotCh <- robot
	robot.send(record)
	//强制进入下一个日志处理器，因为钉钉有可能发送失败
	return true
}

// IsHandling 判断当前处理器是否可以处理日志
func (r *DingTalk) IsHandling(level contract.Level) bool {
	return level <= r.level
}

// Close 关闭日志处理器
func (r *DingTalk) Close() error {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()
	if len(r.robots) == 0 {
		return nil
	}
	for _, robot := range r.robots {
		robot.close()
	}
	r.robots = nil
	return nil
}
