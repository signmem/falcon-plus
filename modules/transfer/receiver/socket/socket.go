package socket

import (
	"github.com/signmem/falcon-plus/modules/transfer/g"
	"log"
	"net"
	"time"
)

func StartSocket() {
	if !g.Config().Socket.Enabled {
		return
	}

	addr := g.Config().Socket.Listen
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("net.ResolveTCPAddr fail: %s", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	} else {
		log.Println("socket listening", addr)
	}

	log.Printf("socket listening on %s success", addr)


	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("listener.Accept error: %v, retry after 100ms", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		go socketTelnetHandle(conn)
	}
}
