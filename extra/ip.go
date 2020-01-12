package extra

import (
	"github.com/buexplain/go-flog"
	"net"
)

//本机ip地址
type IP struct {
	ip string
}

func NewIP() *IP {
	tmp := new(IP)
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.IsLoopback() == false && ipNet.IP.IsGlobalUnicast() == true {
				tmp.ip = ipNet.IP.String()
				break
			}
		}
	}
	return tmp
}

func (this *IP) Processor(record *flog.Record) {
	record.Extra["IP"] = this.ip
}
