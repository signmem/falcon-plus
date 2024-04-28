package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/open-falcon/falcon-plus/common/model"
	"gitlab.tools.vipshop.com/vip-ops-sh/falcon-stats-collector/g"
)

type Counter struct {
	Name string
	Cnt  int64
	Qps  int64
	Time string
}

func unpackServer(server string) (string, string) {
	var endpoint string
	var port string

	if !strings.Contains(server, ":") {
		endpoint = server
		port = ""
	}

	serverStr := strings.Split(server, ":")
	endpoint = serverStr[0]
	port = serverStr[1]

	return endpoint, port
}

func (c *Counter) toTsdbItem(falconType string, metricType string, server string) *TsdbItem {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", c.Time, time.Local)
	if err != nil {
		t = time.Now()
	}

	endpoint, port := unpackServer(server)

	name := fmt.Sprintf("falcon.%s.%s", falconType, c.Name)

	var value float64
	if metricType == "qps" {
		value = float64(c.Qps)
	} else {
		value = float64(c.Cnt)
	}

	tagName := g.Config().Env.Cluster

	return &TsdbItem{
		Metric: name,
		Tags: map[string]string{
			"endpoint": endpoint,
			"port":     port,
			"cluster": tagName,
		},
		Value:     value,
		Timestamp: t.Unix(),
	}
}

// 临时添加，用于追加一份 Counter 数据
func (c *Counter) toTsdbItem4Cnt(falconType string, metricType string, server string) *TsdbItem {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", c.Time, time.Local)
	if err != nil {
		t = time.Now()
	}

	endpoint, port := unpackServer(server)

	// name 做区分
	name := fmt.Sprintf("falcon.%s.%s.c", falconType, c.Name)

	tagName := g.Config().Env.Cluster

	return &TsdbItem{
		Metric: name,
		Tags: map[string]string{
			"endpoint": endpoint,
			"port":     port,
			"cluster": tagName,
		},
		Value:     float64(c.Cnt),
		Timestamp: t.Unix(),
	}
}

type Counters []*Counter

func HealthToTsdbItem(falconType string, server string, health bool) *TsdbItem {
	var value float64
	if health {
		value = 1
	} else {
		value = 0
	}

	endpoint, port := unpackServer(server)
	name := fmt.Sprintf("falcon.%s.health", falconType)
	return &TsdbItem{
		Metric: name,
		Tags: map[string]string{
			"endpoint": endpoint,
			"port":     port,
		},
		Value:     value,
		Timestamp: time.Now().Unix(),
	}
}

type FalconCommon interface {
	Health() (bool, error)
	Version() (string, error)
	Counter() (Counters, error)
	Close()
}

type FalconClient struct {
	url        *url.URL
	httpClient *http.Client
	tr         *http.Transport
}

func NewFalconClient(server string, timeout time.Duration) (*FalconClient, error) {
	u, err := url.Parse(fmt.Sprintf("http://%s", server))
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{}

	return &FalconClient{
		url:        u,
		httpClient: &http.Client{Timeout: timeout},
		tr:         tr,
	}, nil
}

func (c *FalconClient) get(path string) ([]byte, error) {
	u := c.url
	u.Path = path

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to call stats api, status %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func (c *FalconClient) Close() {
	c.tr.CloseIdleConnections()
}

func (c *FalconClient) Health() (bool, error) {
	body, err := c.get("/health")
	if err != nil {
		return false, err
	}

	// 去除空格换行的字符
	body = bytes.TrimSpace(body)
	return string(body) == "ok", nil
}

func (c *FalconClient) Version() (string, error) {
	body, err := c.get("/version")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *FalconClient) Counter() (Counters, error) {
	return nil, nil
}

type TransferCounterResp struct {
	Data Counters `json:"data"`
	Msg  string   `json:"msg"`
}

type FalconTransferClient struct {
	FalconClient
}

func (c *FalconTransferClient) Counter() (Counters, error) {
	body, err := c.get("/counter/all")
	if err != nil {
		return nil, err
	}

	var resp TransferCounterResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

type FalconGraphClient struct {
	FalconClient
}

func (c *FalconGraphClient) Health() (bool, error) {
	body, err := c.get("/api/v2/health")
	if err != nil {
		return false, err
	}

	var resp struct {
		Msg string `json:"msg"`
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return false, err
	}

	return string(resp.Msg) == "ok", nil
}

func (c *FalconGraphClient) Version() (string, error) {
	body, err := c.get("/api/v2/version")
	if err != nil {
		return "", err
	}

	var resp struct {
		Value string `json:"value"`
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}
	return string(resp.Value), nil
}

func (c *FalconGraphClient) Counter() (Counters, error) {
	body, err := c.get("/counter/all")
	if err != nil {
		return nil, err
	}

	var resp Counters
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type KafkaConsumerCounterResp struct {
	Data Counters `json:"data"`
	Msg  string   `json:"msg"`
}

type FalconKafkaConsumerClient struct {
	FalconClient
}

func (c *FalconKafkaConsumerClient) Counter() (Counters, error) {
	body, err := c.get("/counter/all")
	if err != nil {
		return nil, err
	}

	var resp KafkaConsumerCounterResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

type TrendCounterResp struct {
	Data Counters `json:"data"`
	Msg  string   `json:"msg"`
}

type FalconTrendClient struct {
	FalconClient
}

func (c *FalconTrendClient) Counter() (Counters, error) {
	body, err := c.get("/counter/all")
	if err != nil {
		return nil, err
	}

	var resp TrendCounterResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

type FalconJudgeClient struct {
	FalconClient
}

func (c *FalconJudgeClient) Counter() (Counters, error) {
	body, err := c.get("/count")
	if err != nil {
		return nil, err
	}
	countStr := strings.TrimSpace(strings.TrimPrefix(string(body), "total:"))
	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var counters Counters
	counters = append(counters, &Counter{Name: "MetricsCnt", Cnt: count, Qps: 0, Time: time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")})

	return counters, nil

}

func ClientFactory(falconType string, server string, timeout time.Duration) (FalconCommon, error) {
	c, err := NewFalconClient(server, timeout)
	if err != nil {
		return nil, err
	}

	var fc FalconCommon
	switch falconType {
	case "transfer":
		fc = &FalconTransferClient{*c}
	case "graph":
		fc = &FalconGraphClient{*c}
	case "kafka_consumer":
		fc = &FalconKafkaConsumerClient{*c}
	case "trend":
		fc = &FalconTrendClient{*c}
	case "judge":
		fc = &FalconJudgeClient{*c}
	default:
		fc = c
	}

	return fc, nil
}
