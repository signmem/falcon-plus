package kfk

import "fmt"


type MItem struct {
	Metric    string			`json:"metric"`
	Endpoint  string			`json:"endpoint"`
	Step      int64				`json:"step"`
	Type      string      		`json:"counterType"`
	Tags      string			`json:"tags"`
	Value     interface{}		`json:"value"`
	Timestamp int64				`json:"timestamp"`
}

func (this *MItem) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d\t%.3f\t%s\t%d", this.Endpoint,
		this.Metric, this.Tags, this.Timestamp, this.Value, this.Type, this.Step)
}