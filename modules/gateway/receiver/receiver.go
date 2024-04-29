package receiver

import (
	"github.com/signmem/falcon-plus/modules/gateway/receiver/rpc"
	"github.com/signmem/falcon-plus/modules/gateway/receiver/socket"
)

func Start() {
	go rpc.StartRpc()
	go socket.StartSocket()
}
