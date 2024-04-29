package funcs

import (
	"fmt"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/toolkits/nux"
	"github.com/toolkits/slice"
	"log"
)

func PortMetrics() (L []*model.MetricValue) {

	debug := g.Config().Debug
	reportPorts := g.ReportPorts()
	sz := len(reportPorts)
	if sz == 0 {
		return
	}

	allTcpPorts, err := nux.TcpPorts()
	if err != nil {
		log.Println("PortMetrics() error: ", err)
		return
	}

	if debug {
		log.Printf("tcp port : %v", allTcpPorts)
	}

	allUdpPorts, err := nux.UdpPorts()
	if err != nil {
		log.Println("PortMetrics() error: ", err)
		return
	}

	if debug {
		log.Printf("udp port : %v", allUdpPorts)
	}

	for i := 0; i < sz; i++ {
		tags := fmt.Sprintf("port=%d", reportPorts[i])
		if slice.ContainsInt64(allTcpPorts, reportPorts[i]) || slice.ContainsInt64(allUdpPorts, reportPorts[i]) {
			L = append(L, GaugeValue(g.NET_PORT_LISTEN, 1, tags))
		} else {
			L = append(L, GaugeValue(g.NET_PORT_LISTEN, 0, tags))
		}
	}

	return
}
