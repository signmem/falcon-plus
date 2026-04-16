package funcs

import (
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/signmem/nux"
	"log"
)

func MemMetrics() []*model.MetricValue {
	m, err := nux.MemInfo()
	if err != nil {
		log.Printf("[ERROR] MemMetrics() error: %s\n", err)
		return nil
	}

	memBuffer := m.Buffers
	memCache  := m.Cached
	memFree := m.MemFree + m.Buffers + m.Cached
	memUsed := m.MemTotal - memFree
	memAvaril := m.MemAvailable
	memSlab := m.Slab
	memSlabReclaimable := m.Reclaimable
	memSlabUnreclaim  := m.Unreclaim
	var memSlabPercent float64

	pmemFree := 0.0
	pmemUsed := 0.0
	memSlabPercent = 0.0

	if m.MemTotal != 0 {
		pmemFree = float64(memFree) * 100.0 / float64(m.MemTotal)
		pmemUsed = float64(memUsed) * 100.0 / float64(m.MemTotal)
		memSlabPercent =  ( float64(memSlab) / float64(m.MemTotal) )* 100
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
		GaugeValue("mem.available", memAvaril),
		GaugeValue("mem.cache", memCache),
		GaugeValue("mem.buffer", memBuffer),
		GaugeValue("mem.swaptotal", m.SwapTotal),
		GaugeValue("mem.swapused", m.SwapUsed),
		GaugeValue("mem.swapfree", m.SwapFree),
		GaugeValue("mem.memfree.percent", pmemFree),
		GaugeValue("mem.memused.percent", pmemUsed),
		GaugeValue("mem.swapfree.percent", pswapFree),
		GaugeValue("mem.swapused.percent", pswapUsed),
		GaugeValue("mem.slab.size", memSlab),
		GaugeValue("mem.slab.reclaimable", memSlabReclaimable),
		GaugeValue("mem.slab.unreclaim", memSlabUnreclaim),
		GaugeValue("mem.slab.percent", memSlabPercent),
	}

}
