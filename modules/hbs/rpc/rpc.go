package rpc

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync/atomic"
	"time"

	"github.com/signmem/falcon-plus/modules/hbs/g"
)

type Hbs int
type Agent int

type ConnMonitor struct {
	ActiveConns int64
	TotalConns  int64
	MaxConns    int64
	StartTime   time.Time
}

var connMonitor = &ConnMonitor{
	StartTime: time.Now(),
}

type MonitoredConn struct {
	net.Conn
}

func (mc *MonitoredConn) Close() error {
	atomic.AddInt64(&connMonitor.ActiveConns, -1)
	return mc.Conn.Close()
}

func newMonitoredConn(conn net.Conn) *MonitoredConn {
	atomic.AddInt64(&connMonitor.ActiveConns, 1)
	atomic.AddInt64(&connMonitor.TotalConns, 1)

	currentConns := atomic.LoadInt64(&connMonitor.ActiveConns)
	maxConns := atomic.LoadInt64(&connMonitor.MaxConns)
	if currentConns > maxConns {
		atomic.StoreInt64(&connMonitor.MaxConns, currentConns)
	}

	if currentConns%100 == 0 {
		log.Printf("[MONITOR] Current: %d, Total: %d",
			currentConns, atomic.LoadInt64(&connMonitor.TotalConns))
	}

	return &MonitoredConn{Conn: conn}
}

func GetConnStats() (activeConns, totalConns, maxConns int64) {
	return atomic.LoadInt64(&connMonitor.ActiveConns),
		atomic.LoadInt64(&connMonitor.TotalConns),
		atomic.LoadInt64(&connMonitor.MaxConns)
}

func CheckBottleneck(connection int64 ) bool {
	activeConns := atomic.LoadInt64(&connMonitor.ActiveConns)

	if activeConns > connection {
		return true
	}
	return false
}


func Start() {
	addr := g.Config().Listen

	server := rpc.NewServer()
	// server.Register(new(filter.Filter))
	server.Register(new(Agent))
	server.Register(new(Hbs))

	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatalln("listen error:", e)
	} else {
		log.Println("listening", addr)
	}

	go startMonitorTicker()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("listener accept fail:", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}

		monitoredConn := newMonitoredConn(conn)
		go server.ServeCodec(jsonrpc.NewServerCodec(monitoredConn))
		// go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}


func startMonitorTicker() {

	ticker := time.NewTicker(360 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		activeConns, totalConns, maxConns := GetConnStats()
		uptime := time.Since(connMonitor.StartTime)

		log.Printf("[MONITOR] runtime: %v, active: %d, total: %d, max: %d",
			uptime, activeConns, totalConns, maxConns)

		if CheckBottleneck(10000) {
			log.Printf("[WARNING] too much connetion now: %d ", activeConns)
		}
	}
}