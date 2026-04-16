package falcon

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/cmdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/net"
	"github.com/signmem/falcon-plus/modules/pingcheck/selector"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"sync"
	"time"
)

var (
	FalconPingMinMetrics []*MetricValue
	metricsMu            sync.RWMutex
)

const (
	masterCheckInterval = 1 * time.Second
	emptyHostSleep = 60 * time.Second
	taskSleepInterval = 10 * time.Second
	pingMetricStep = 60
)

func FalconPingMetrics() {

	time.Sleep(taskSleepInterval)
	g.Logger.Infof("[FalconPing] module started, role: %s", selector.Role)

	falconPingConcurrent := g.Config().FalconPingConcurrent

	for {

		// slave == testing
		// master == production
		// must be modify  terry.zeng 2028-03-18

		if selector.Role != "master" {
			time.Sleep(masterCheckInterval)
			continue
		}

		checkPing := cmdb.RePingMetrics
		totalHosts := len(checkPing)

		if totalHosts == 0 {

			if g.Config().Debug {
				g.Logger.Debug("[FalconPing] host check Zero sleep 60s")
			}

			time.Sleep(emptyHostSleep)
			continue
		}


		g.Logger.Debugf("[FalconPing check Start] ping host total %d", totalHosts )

		g.FalconPingCheckTotal.Set(totalHosts)

		metricsMu.Lock()
		FalconPingMinMetrics = make([]*MetricValue, 0, totalHosts)
		metricsMu.Unlock()

		taskChan := make(chan struct{}, falconPingConcurrent)
		var wg sync.WaitGroup
		resultChan := make(chan *MetricValue, totalHosts)

		for i := range checkPing {
			host := checkPing[i]
			taskChan <- struct{}{}
			wg.Add(1)

			go func(hh cmdb.HostInfo) {
				defer func() {
					<-taskChan
					wg.Done()
				}()

				falconPingHostMakeMetrics(hh, resultChan)

			}(host)
		}

		go func() {
			wg.Wait()
			close(resultChan)
			close(taskChan)
		}()


		var tempMetrics []*MetricValue
		for res := range resultChan {
			tempMetrics = append(tempMetrics, res)
			if res.Value == 0 {
				g.FalconPingCheckFailuer.Incr()
			} else {
				g.FalconPingCheckSuccess.Incr()
			}
		}

		metricsMu.Lock()

		finishCount := len(tempMetrics)
		metricsMu.Unlock()

		g.Logger.Debugf("[FalconPing] check finished, success count: %d", finishCount)


		// add agent.alive to agent.ping

		agentAliveMetrics := AddAgentAliveToPing(cmdb.AgentAlive)
		FalconPingMinMetrics = append(tempMetrics, agentAliveMetrics...)

		if g.Config().Debug {
			g.Logger.Debugf("[Falcon Send agent.ping to transfer] Total: %d", len(FalconPingMinMetrics))
		}
		/*
		if err := sendMetric(FalconPingMinMetrics); err != nil {
			g.Logger.Errorf("[FalconPing] send metrics failed: %v", err)
		}
		*/

		SendToTransfer(FalconPingMinMetrics)

		time.Sleep(taskSleepInterval)
	}

}


func falconPingHostMakeMetrics(host cmdb.HostInfo, resultChan chan<- *MetricValue) {

	// 用于检测主机主机是否可 ping

	pingMetrics := &MetricValue{}

	pingMetrics.Type = "GAUGE"
	pingMetrics.Step = pingMetricStep
	pingMetrics.Metric = "agent.ping"
	pingMetrics.Timestamp  = time.Now().Unix()
	pingMetrics.Endpoint = host.HostName

	pingStatus, _ := net.PingFromProxy(host.IPAddr)

	if pingStatus == false {
		pingMetrics.Value = 0
	} else {
		pingMetrics.Value = 1
	}

    /*
	if g.Config().Debug == true {
		g.Logger.Debugf("[FalconCheck] Metrics: %s" , pingMetrics.String())
	}
    */

	resultChan <- pingMetrics
}

func AddAgentAliveToPing(agentAliveHost []cmdb.HostInfo) (pingMetrics []*MetricValue) {

	for _, host := range agentAliveHost {

		hostMetrics := &MetricValue{}
		hostMetrics.Timestamp  = time.Now().Unix()
		hostMetrics.Type       = "GAUGE"
		hostMetrics.Step       = pingMetricStep
		hostMetrics.Metric     = "agent.ping"
		hostMetrics.Timestamp  = time.Now().Unix()
		hostMetrics.Value      = 1
		hostMetrics.Endpoint = host.HostName

		pingMetrics = append(pingMetrics, hostMetrics)
	}

	return pingMetrics
}