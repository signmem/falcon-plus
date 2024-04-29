package net

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"net"
	"time"
)

func SshCheck(host string) bool {
	port := "22"
	timeout := time.Second
	_, err := net.DialTimeout("tcp", host + ":" + port  , timeout)
	if err != nil {
		_, err_msg := err.Error()[0], err.Error()[5:]
		g.Logger.Errorf("SshCheck() ssh Check Error host: %s, err: %s", host, err_msg)
		return false
	}
	return true
}

