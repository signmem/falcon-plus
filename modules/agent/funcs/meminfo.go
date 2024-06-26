package funcs

import (
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/toolkits/nux"
	"log"
)

func MemMetrics() []*model.MetricValue {
	m, err := nux.MemInfo()
	if err != nil {
		log.Println("MemMetrics() error", err)
		return nil
	}

	memFree := m.MemFree + m.Buffers + m.Cached
	memUsed := m.MemTotal - memFree

	pmemFree := 0.0
	pmemUsed := 0.0
	if m.MemTotal != 0 {
		pmemFree = float64(memFree) * 100.0 / float64(m.MemTotal)
		pmemUsed = float64(memUsed) * 100.0 / float64(m.MemTotal)
	}

	pswapFree := 0.0
	pswapUsed := 0.0
	if m.SwapTotal != 0 {
		pswapFree = float64(m.SwapFree) * 100.0 / float64(m.SwapTotal)
		pswapUsed = float64(m.SwapUsed) * 100.0 / float64(m.SwapTotal)
	}

	if g.SkipSwapMonitor {
		m.SwapTotal = 4194304000
		m.SwapUsed = 1000
		m.SwapFree = 4194303000
		pswapFree = 99.00
		pswapUsed = 1.00
	}

	return []*model.MetricValue{
		GaugeValue("mem.memtotal", m.MemTotal),
		GaugeValue("mem.memused", memUsed),
		GaugeValue("mem.memfree", memFree),
		GaugeValue("mem.swaptotal", m.SwapTotal),
		GaugeValue("mem.swapused", m.SwapUsed),
		GaugeValue("mem.swapfree", m.SwapFree),
		GaugeValue("mem.memfree.percent", pmemFree),
		GaugeValue("mem.memused.percent", pmemUsed),
		GaugeValue("mem.swapfree.percent", pswapFree),
		GaugeValue("mem.swapused.percent", pswapUsed),
	}

}
