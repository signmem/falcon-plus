package g

import "time"

func NewSCount(name string) *SCount {
	uts := time.Now().Unix()
	return &SCount{Name: name, Cnt: 0, Time: uts}
}

func (this *SCount) Get() *SCount {
	this.RLock()
	uts := time.Now().Unix()
	defer this.RUnlock()

	return &SCount{
		Name: this.Name,
		Cnt: this.Cnt,
		Time: uts,
	}
}

func (this *SCount) Incr() {
	this.Lock()
	this.Cnt += int64(1)
	this.Unlock()
}

func (this *SCount) Set(num int) {
	this.Lock()
	this.Cnt = int64(num)
	this.Unlock()
}

func (this *SCount) Reset() {
	this.Lock()
	this.Cnt = int64(0)
	this.Unlock()
}