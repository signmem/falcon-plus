package falcon

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"math/rand"
	"sync"
	"time"
)

var (
	TransferClientsLock *sync.RWMutex                   = new(sync.RWMutex)
	TransferClients     map[string]*SingleConnRpcClient = map[string]*SingleConnRpcClient{}
)


func SendToTransfer(metrics []*MetricValue) {
	if len(metrics) == 0 {
		return
	}

	var resp TransferResponse
	SendMetrics(metrics, &resp)
}



func SendMetrics(metrics []*MetricValue, resp *TransferResponse) {

	rand.Seed(time.Now().UnixNano())
	for _, i := range rand.Perm(len(g.Config().Transfer.Addrs)) {
		addr := g.Config().Transfer.Addrs[i]

		c := getTransferClient(addr)
		if c == nil {
			c = initTransferClient(addr)
		}

		if updateMetrics(c, metrics, resp) {
			break
		}
	}
}

func updateMetrics(c *SingleConnRpcClient, metrics []*MetricValue, resp *TransferResponse) bool {
	err := c.Call("Transfer.Update", metrics, resp)
	if err != nil {
		g.Logger.Error("call Transfer.Update fail:", c, err)
		return false
	}
	return true
}

func initTransferClient(addr string) *SingleConnRpcClient {
	var c *SingleConnRpcClient = &SingleConnRpcClient{
		RpcServer: addr,
		Timeout:   time.Duration(g.Config().Transfer.Timeout) * time.Millisecond,
	}
	TransferClientsLock.Lock()
	defer TransferClientsLock.Unlock()
	TransferClients[addr] = c

	return c
}

func getTransferClient(addr string) *SingleConnRpcClient {
	TransferClientsLock.RLock()
	defer TransferClientsLock.RUnlock()

	if c, ok := TransferClients[addr]; ok {
		return c
	}
	return nil
}
