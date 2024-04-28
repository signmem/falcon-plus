package rpc

import (
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/trend/cache"
)

type Trend int

func (this *Trend) Ping(req model.NullRpcRequest, resp *model.SimpleRpcResponse) error {
	return nil
}

func (this *Trend) Send(items []*model.TrendItem, resp *model.SimpleRpcResponse) error {
	go cache.Push(items)
	return nil
}
