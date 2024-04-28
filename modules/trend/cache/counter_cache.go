package cache

import (

	"sync"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
)

type CounterItem struct {
	sync.RWMutex
	Endpoint  string
	Metric    string
	Tags      string
	DsType    string
	Step      int
	LastValue float64
	LastTime  int64
	Max       float64
	Min       float64
	Sum       float64
	Num       int
}

func (this *CounterItem) Update(val *model.TrendItem) {
	this.Lock()
	defer this.Unlock()
	if val == nil {
		g.Logger.Error("Struct CounterItem Update() parameter is nil.")
		return
	}
	value := val.Value - this.LastValue
	duration := val.Timestamp - this.LastTime
	//迟来的数据丢弃
	if duration < 0 {
		return
	}
	this.LastValue = val.Value
	this.LastTime = val.Timestamp
	if value >= 0 { //只处理增长的counter，以去除重启导致的计数值重置
		v := value / float64(duration)
		this.Step = val.Step
		if this.Num <= 0 {
			this.Max = v
			this.Min = v
		} else {
			if v > this.Max {
				this.Max = v
			} else if v < this.Min {
				this.Min = v
			}
		}
		this.Sum += v
		this.Num++
	}
	//log.Debugf("update counter item:[%v]", this)
}

func (this *CounterItem) GetTrendResult() *g.TrendResult {
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

func CounterNew(key int64, pk string, val *model.TrendItem) interface{} {
	if val == nil {
		g.Logger.Error("CounterNew() input parameter is nil.")
		return nil
	}
	if CounterCacheHistory != nil {
		item, ok := CounterCacheHistory.GetItem(key-1, pk)
		if ok && item != nil {
			prev_item := item.(*CounterItem)
			duration := val.Timestamp - prev_item.LastTime
			value := val.Value - prev_item.LastValue
			if duration < int64(2*val.Step) && value >= 0 {
				//value := val.Value - prev_item.LastValue
				v := value / float64(duration)
				return &CounterItem{
					Endpoint:  val.Endpoint,
					Metric:    val.Metric,
					Tags:      val.Tags,
					DsType:    val.DsType,
					Step:      val.Step,
					LastValue: val.Value,
					LastTime:  val.Timestamp,
					Max:       v,
					Min:       v,
					Sum:       v,
					Num:       1,
				}
			}
		}
	}
	return &CounterItem{
		Endpoint:  val.Endpoint,
		Metric:    val.Metric,
		Tags:      val.Tags,
		DsType:    val.DsType,
		Step:      val.Step,
		LastValue: val.Value,
		LastTime:  val.Timestamp,
		Num:       0,
	}
}

func CounterUpdate(item interface{}, val *model.TrendItem) {
	counter_item := item.(*CounterItem)
	if counter_item == nil {
		g.Logger.Error("CounterUpdate() convert to *CounterItem is nil.")
		return
	}
	counter_item.Update(val)
}

func CounterTrendResult(item interface{}) *g.TrendResult {
	counter_item := item.(*CounterItem)
	if counter_item == nil {
		g.Logger.Error("CounterTrendResult() convert to *CounterItem is nil.")
		return nil
	}
	return counter_item.GetTrendResult()
}
