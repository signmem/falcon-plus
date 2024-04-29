package model

import (
	"fmt"

	"github.com/signmem/falcon-plus/common/utils"
)

type TrendItem struct {
	Endpoint  string  `json:"endpoint"`
	Metric    string  `json:"metric"`
	Tags      string  `json:"tags"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	DsType    string  `json:"dstype"`
	Step      int     `json:"step"`
}

func (this *TrendItem) PrimaryKey() string {
	return utils.Md5(this.PK())
}

func (this *TrendItem) PK() string {
	var pk string
	if this.Tags == "" {
		pk = fmt.Sprintf("%s/%s", this.Endpoint, this.Metric)
	} else {
		pk = fmt.Sprintf("%s/%s/%s", this.Endpoint, this.Metric, this.Tags)
	}
	return pk
}
