package rpc

import (
	"container/list"
	"net"
	"net/rpc"
	"sync"
	"time"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
)

type conn_list struct {
	sync.RWMutex
	list *list.List
}

func (l *conn_list) insert(c net.Conn) *list.Element {
	l.Lock()
	defer l.Unlock()
	return l.list.PushBack(c)
}
func (l *conn_list) remove(e *list.Element) net.Conn {
	l.Lock()
	defer l.Unlock()
	return l.list.Remove(e).(net.Conn)
}

var Close_chan, Close_done_chan chan int
var connects conn_list

func init() {
	Close_chan = make(chan int, 1)
	Close_done_chan = make(chan int, 1)
	connects = conn_list{list: list.New()}
}

func Start() {
	if !g.Config().Rpc.Enabled {
		g.Logger.Warning("rpc.Start warning, not enabled")
		return
	}
	addr := g.Config().Rpc.Listen
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		g.Logger.Errorf("rpc.Start error, net.ResolveTCPAddr failed, %s", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		g.Logger.Errorf("rpc.Start error, listen %s failed, %s", addr, err)
	} else {
		g.Logger.Infof("rpc.Start ok, listening on: %s", addr)
	}

	rpc.Register(new(Trend))

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				g.Logger.Errorf("listener.Accept occur error: %s", err)
				continue
			}
			go func() {
				e := connects.insert(conn)
				defer connects.remove(e)
				rpc.ServeConn(conn)
			}()
		}
	}()

	select {
	case <-Close_chan:
		g.Logger.Errorf("rpc, recv sigout and exiting...")
		listener.Close()
		Close_done_chan <- 1

		connects.Lock()
		for e := connects.list.Front(); e != nil; e = e.Next() {
			e.Value.(net.Conn).Close()
		}
		connects.Unlock()

		return
	}

}
