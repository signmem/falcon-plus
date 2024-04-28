package net

import (
	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"github.com/tatsushid/go-fastping"
	"github.com/open-falcon/falcon-plus/common/http"
	"io/ioutil"
	"net"
	"os/exec"
	"time"
	"encoding/json"
)


func PingStatus(ip string) bool {
	cmd := exec.Command("ping", ip, "-c", "2", "-W", "3")
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func PingFromProxy(ip string) bool {
	// terry.zeng

	pingProxyServer := g.Config().Proxy.Servers

	var pingRequest g.HttpPingRequest
	pingRequest.Ipaddr = ip

	jsonIpInfo, _ := json.Marshal(pingRequest)

	for _, server := range pingProxyServer {

		var httpRespon g.HttpPingResponse

		api := "/api/v1/pingcheck"
		url := "http://" + server + api
		response, err := http.HttpApiPost(url, jsonIpInfo,"")
		if err != nil {
			g.Logger.Warningf("http access %s error:%s", url, err)
			continue
		}

		responseBody, err := ioutil.ReadAll(response)
		_  = json.Unmarshal(responseBody, &httpRespon)

		if httpRespon.PingStatus == true {
			return true
		}

	}

	return false
}


func CheckPing(ip string) (status bool, err error) {

	pingStatus := false
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)

	if err != nil {
		g.Logger.Errorf("CheckPing() net.ResolveIPAddr() error:%s", err)
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
