package cache

import (
	"sync"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
)

type GaugeItem struct {
	sync.RWMutex
	Endpoint string
	Metric   string
	Tags     string
	DsType   string
	Step     int
	Max      float64
	Min      float64
	Sum      float64
	Num      int
}

func (this *GaugeItem) Update(val *model.TrendItem) {
	this.Lock()
	defer this.Unlock()
	if val == nil {
		g.Logger.Error("Struct GaugeItem Update() input parameter is nil")
		return
	}
	this.Step = val.Step
	if val.Value > this.Max {
		this.Max = val.Value
	} else if val.Value < this.Min {
		this.Min = val.Value
	}
	this.Sum += val.Value
	this.Num++
}

func (this *GaugeItem) GetTrendResult() *g.TrendResult {
	this.RLock()
	defer this.RUnlock()
	if this.Num <= 0 {
		return nil
	}
	return &g.TrendResult{
		Endpoint: this.Endpoint,
		Metric:   this.Metric,
		Tags:     this.Tags,
		DsType:   this.DsType,
		Step:     this.Step,
		Max:      this.Max,
		Min:      this.Min,
		Avg:      this.Sum / float64(this.Num),
		Num:      this.Num,
	}
}

func GuageNew(key int64, pk string, val *model.TrendItem) interface{} {
	if val == nil {
		g.Logger.Error("GuageNew() input parameter is nil")
		return nil
	}
	return &GaugeItem{
		Endpoint: val.Endpoint,
		Metric:   val.Metric,
		Tags:     val.Tags,
		DsType:   val.DsType,
		Step:     val.Step,
		Max:      val.Value,
		Min:      val.Value,
		Sum:      val.Value,
		Num:      1,
	}
}

func GuageUpdate(item interface{}, val *model.TrendItem) {
	guage_item := item.(*GaugeItem)
	if guage_item == nil {
		g.Logger.Error("GuageUpdate() convert to *GaugeItem is nil.")
		return
	}
	guage_item.Update(val)
}

func GuageTrendResult(item interface{}) *g.TrendResult {
	guage_item := item.(*GaugeItem)
	if guage_item == nil {
		g.Logger.Error("GuageTrendResult() convert to *GaugeItem is nil.")
		return nil
	}
	return guage_item.GetTrendResult()
}
