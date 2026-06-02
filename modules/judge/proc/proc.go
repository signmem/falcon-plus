package proc

import (
	"sync"
	"time"
)

var (

	SendToRedisCntTotal     =  NewSCount("SendToRedisCntTotal")
	SendToRedisCntSuccess   =  NewSCount("SendToRedisCntSuccess")
	SendToRedisCntDrop      =  NewSCount("SendToRedisCntDrop")
)


type SCount struct {
	sync.RWMutex
	Name    string			`json:"Name"`
	Cnt     float64			`json:"Cnt"`
	Qps	float64			`json:"Qps"`
	Time    string			`json:"Time"`
	Other   map[string]interface{}  `json:"Other"`
}

func NewSCount(name string) *SCount {
	t := time.Now()
	uts := t.Format("2006-01-02 15:04:05")
	return &SCount{Name: name, Cnt: 0, Qps: 0, Time: uts, Other: make(map[string]interface{}) } 
}

func (this *SCount) Get() *SCount {
	this.RLock()
	t := time.Now()
	uts := t.Format("2006-01-02 15:04:05")
	defer this.RUnlock()

	return &SCount{
		Name: this.Name,
		Qps: 0,
		Cnt: this.Cnt,
		Time: uts,
		Other:  make(map[string]interface{}),
	}
}

func (this *SCount) Incr() {
	this.Lock()
	this.Cnt += float64(1)
	this.Unlock()
}

func  GetAll() []interface{} {
	return []interface{}{
		SendToRedisCntTotal.Get(),
		SendToRedisCntDrop.Get(),
		SendToRedisCntSuccess.Get(),
	}
}


func GenMap(name string, value int)  *SCount { 
	t := time.Now()
	uts := t.Format("2006-01-02 15:04:05")

	return &SCount{Name: name, Cnt: float64(value), Qps: 0,  Time: uts,  Other: make(map[string]interface{}) }
}
