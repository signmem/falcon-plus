package falcon

import (
	"fmt"
	"github.com/signmem/netlib"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"math"
	"net/rpc"
	"sync"
	"time"
)


type MetricValue struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}


func (this *MetricValue) String() string {
	return fmt.Sprintf(
		"<Endpoint:%s, Metric:%s, Type:%s, Tags:%s, Step:%d, Time:%d, Value:%v>",
		this.Endpoint,
		this.Metric,
		this.Type,
		this.Tags,
		this.Step,
		this.Timestamp,
		this.Value,
	)
}



type TransferResponse struct {
	Message string
	Total   int
	Invalid int
	Latency int64
}

func (this *TransferResponse) String() string {
	return fmt.Sprintf(
		"<Total=%v, Invalid:%v, Latency=%vms, Message:%s>",
		this.Total,
		this.Invalid,
		this.Latency,
		this.Message,
	)
}


type SingleConnRpcClient struct {
	sync.Mutex
	rpcClient *rpc.Client
	RpcServer string
	Timeout   time.Duration
}

func (this *SingleConnRpcClient) close() {
	if this.rpcClient != nil {
		this.rpcClient.Close()
		this.rpcClient = nil
	}
}

func (this *SingleConnRpcClient) serverConn() error {
	if this.rpcClient != nil {
		return nil
	}

	var err error
	var retry int = 1

	for {
		if this.rpcClient != nil {
			return nil
		}

		this.rpcClient, err = netlib.JsonRpcClient("tcp", this.RpcServer, this.Timeout)
		if err != nil {

			g.Logger.Errorf("dial %s fail: %v\n", this.RpcServer, err)

			if retry > 3 {
				return err
			}
			time.Sleep(time.Duration(math.Pow(2.0, float64(retry))) * time.Second)
			retry++
			continue
		}
		return err
	}
}

func (this *SingleConnRpcClient) Call(method string, args interface{}, reply interface{}) error {

	this.Lock()
	defer this.Unlock()

	err := this.serverConn()
	if err != nil {
		return err
	}

	timeout := time.Duration(10 * time.Second)
	done := make(chan error, 1)

	go func() {
		err := this.rpcClient.Call(method, args, reply)
		done <- err
	}()

	select {
	case <-time.After(timeout):
		g.Logger.Errorf("[WARN] rpc call timeout %v => %v", this.rpcClient, this.RpcServer)
		this.close()
	case err := <-done:
		if err != nil {
			this.close()
			return err
		}
	}

	return nil
}
