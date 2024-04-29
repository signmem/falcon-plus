package receiver

import (
	"github.com/signmem/falcon-plus/modules/transfer/receiver/rpc"
	"github.com/signmem/falcon-plus/modules/transfer/receiver/socket"
)

func Start() {
	go rpc.StartRpc()
	go socket.StartSocket()
}
