package extra

import (
	"github.com/buexplain/go-flog/contract"
	"net"
)

// IP 本机ip地址
type IP string

func NewIP() IP {
	var tmp IP
	if address, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range address {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.IsLoopback() == false && ipNet.IP.IsGlobalUnicast() == true {
				tmp = IP(ipNet.IP.String())
				break
			}
		}
	}
	return tmp
}

func (r IP) Processor(record *contract.Record) {
	record.Extra["IP"] = string(r)
}
