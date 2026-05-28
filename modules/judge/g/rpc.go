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
			// this.rpcClient, err = rpc.DialTimeout("tcp", s, this.Timeout)
			if err == nil {
				log.Printf("connect hbs %s success", s)
				return
			}
			log.Printf("dial hbs %s fail: %v", s, err)
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

	timeout := time.After(this.CallTimeout)
	done := make(chan error, 1)

	go func() {
		done <- client.Call(method, args, reply)
	}()

	select {
	case <-timeout:
		this.close()
		return errors.New("call hbs timeout")
	case err := <-done:
		if err != nil {
			this.close()
		}
		return err
	}


	/*
	//  1. 建立连接
	this.insureConn()

	// 核心修复：连接失败，直接返回错误，避免空指针

	done := make(chan error, 1)

	go func() {
		// 现在这里绝对安全，不会 panic
		if this.rpcClient == nil {
			done <- errors.New("rpc client is nil")
			return
		}
		done <- this.rpcClient.Call(method, args, reply)
	}()

	var err error

	select {
	case <-time.After(this.CallTimeout):
		err = errors.New("call hbs timeout")
	case err = <-done:
	}

	if err != nil {
		this.close()
	}

	return err
	*/
}
