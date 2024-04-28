package g

import "fmt"

type TrendResult struct {
	Endpoint string
	Metric   string
	Tags     string
	DsType   string
	Step     int
	Max      float64
	Min      float64
	Avg      float64
	Num      int
}

func (t *TrendResult) String() string {
	return fmt.Sprintf("Endpoint:%s, Metric:%s, Tags:%s, DsType:%s," +
		" Step:%d, Max:%f, Min:%f, Avg:%f, Num:%d", t.Endpoint, t.Metric, t.Tags,
		t.DsType, t.Step, t.Max, t.Min, t.Avg, t.Num)
}