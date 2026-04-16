package net

import (
	"bytes"
	"encoding/json"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"io"
	"net/http"
	"time"
	"context"
	"errors"
)


var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     30 * time.Second,
	},
}

func PingFromProxy(ip string) (status bool, err error) {
	// terry.zeng

	pingProxyServer := g.Config().Proxy.Servers

	pingRequest := g.HttpPingRequest{Ipaddr: ip}

	jsonIpInfo, _ := json.Marshal(pingRequest)

	for _, server := range pingProxyServer {

		var httpRespon g.HttpPingResponse

		api := "/api/v1/pingcheck"
		url := "http://" + server + api
		response, err := httpPost(url, jsonIpInfo)

		if err != nil {
			g.Logger.Warningf("%s http ping response %s error:%s", server, ip, err)
			continue
		}

		_  = json.Unmarshal(response, &httpRespon)

		if httpRespon.PingStatus == true {
			return true, nil
		}

	}

	return false, nil
}

/*
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
*/

func httpPost(fullApiUrl string, params []byte) ([]byte, error) {
	// use to access http post
	// params = post params  [must be []byte format]
	// return http response


	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", fullApiUrl, bytes.NewBuffer(params))

	if err != nil {
		return nil, errors.New("HttpApiPost() http post error with NewRequest")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()


	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return body, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("HttpApiPost() resp status code not 200.")
	}
	return body, nil
}