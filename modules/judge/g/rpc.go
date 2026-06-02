package g

import (
	"errors"
	"github.com/toolkits/net"
	"log"
	"math"
	"net/rpc"
	"sync"
	"time"
)

type SingleConnRpcClient struct {
	sync.RWMutex
	rpcClient   *rpc.Client
	RpcServers  []string
	Timeout     time.Duration
	CallTimeout time.Duration
}

func (this *SingleConnRpcClient) close() {
	if this.rpcClient != nil {
		this.rpcClient.Close()
		this.rpcClient = nil
	}
}

func (this *SingleConnRpcClient) insureConn() {
	this.Lock()
	defer this.Unlock()

	if this.rpcClient != nil {
		return
	}

	var err error
	var retry int = 1

	for {
		if this.rpcClient != nil {
			return
		}

		for _, s := range this.RpcServers {
			this.rpcClient, err = net.JsonRpcClient("tcp", s, this.Timeout)
			if err == nil {
				log.Printf("[INFO] insureConn connect hbs server %s success", s)
				return
			}
			log.Printf("[ERROR] insureConn dial hbs %s fail: %v", s, err)
		}

		if retry > 6 {
			retry = 1
		}

		time.Sleep(time.Duration(math.Pow(2.0, float64(retry))) * time.Second)

		retry++
	}
}

func (this *SingleConnRpcClient) Call(method string, args interface{}, reply interface{}) error {

	this.insureConn()

	this.Lock()
	client := this.rpcClient
	this.Unlock()

	if client == nil {
		return errors.New("rpc client not ready")
	}

	log.Printf("[INFO] hbs connect success, start call: %s", method)

	timeout := time.After(this.CallTimeout)
	done := make(chan error, 1)

	go func() {
		done <- client.Call(method, args, reply)
	}()

	select {
	case <-timeout:
		log.Printf("[ERROR] call hbs %s timeout", method)
		this.close()
		return errors.New("call hbs timeout")
	case err := <-done:
		if err != nil {
			log.Printf("[ERROR] call hbs %s fail: %v", method, err)
			this.close()
		} else {
			log.Printf("[INFO] call hbs %s success", method)
		}

		return err
	}
}
