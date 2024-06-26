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
	sync.Mutex
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
			if err == nil {
				return
			}

			log.Printf("dial %s fail: %s", s, err)
		}

		if retry > 6 {
			retry = 1
		}

		time.Sleep(time.Duration(math.Pow(2.0, float64(retry))) * time.Second)

		retry++
	}
}

func (this *SingleConnRpcClient) Call(method string, args interface{}, reply interface{}) error {

	this.Lock()
	defer this.Unlock()

	this.insureConn()

	done := make(chan error, 1)
	go func() {
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
}
