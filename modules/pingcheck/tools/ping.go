package tools

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
)

func CheckPing(ip string) (status bool, err error) {

	pingStatus := false
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)

	if err != nil {
		g.Logger.Errorf("CheckPing() ResolveIPAddr() error:%s", err)
		return false, err
	}

	p.AddIPAddr(ra)

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		pingStatus = true
	}

	err = p.Run()

	if err != nil {
		g.Logger.Errorf("CheckPing() Run() error:%s", err)
		return false, err
	}

	return pingStatus, nil
}
