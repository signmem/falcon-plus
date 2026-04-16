package falcon

import (
	"bytes"
	"fmt"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"net/http"
	"os"
	"time"
	"encoding/json"
)

var HOSTNAME string

func UploadMetric() {

	if len(HOSTNAME) == 0 {
		HOSTNAME, _ = os.Hostname()
	}

	step := g.Config().FalconAgent.Step

	for {
		totalMetric := createMetrics()
		sendMetric(totalMetric)
		time.Sleep(time.Duration(step) * time.Second)
	}
}

func sendMetric(metrics []MetricValue) (err error) {

	addr := g.Config().FalconAgent.Address
	port := g.Config().FalconAgent.Port

	api := "/v1/push"
	url := "http://" + addr + ":" + port + api

	jdata, err := json.Marshal(metrics)

	if err != nil {
		g.Logger.Errorf("sendMetric() err:%s", err)
		return err
	}

	header := "application/json"
	resp, err := http.Post(url, header, bytes.NewBuffer(jdata))

	if err != nil {
		g.Logger.Errorf("sendMetric() err:%s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("sendMetric() http resp.status err %d", resp.StatusCode)
		g.Logger.Errorf(msg)
	}

	return nil

}

func createMetrics() (sendMetrics []MetricValue) {

	var totalSCountGauge []*g.SCount
	var totalSCountCounter []*g.SCount
	var sendMetricsGauge []MetricValue
	var sendMetricsCounter []MetricValue

	totalSCountGauge = append(totalSCountGauge, g.FalconAgentAliveReportSuccess)
	totalSCountGauge = append(totalSCountGauge, g.FalconAgentAliveAlarmCounter)
	totalSCountGauge = append(totalSCountGauge, g.FalconAgentAliveReportTimeOut)
	totalSCountGauge = append(totalSCountGauge, g.FalconAgentCMDBReportFailure)
	totalSCountGauge = append(totalSCountGauge, g.FalconCMDBHostNameReplicate)
	totalSCountGauge = append(totalSCountGauge, g.FalconPingDie)
	totalSCountGauge = append(totalSCountGauge, g.FalconPingTotal)
	totalSCountGauge = append(totalSCountGauge, g.FalconPingCheckTotal)

	for _, scount := range totalSCountGauge {
		metricType := "GAUGE"
		metric := transferMetrics(scount, metricType)
		sendMetricsGauge = append(sendMetricsGauge, metric)
	}

	totalSCountCounter = append(totalSCountCounter, g.FalconAgentAliveAlarmSuccess)
	totalSCountCounter = append(totalSCountCounter, g.FalconAgentAliveAlarmFailure)
	totalSCountCounter = append(totalSCountCounter, g.FalconAgentAliveAlarmDegarded)
	totalSCountCounter = append(totalSCountCounter, g.FalconCMDBHostNameAlarmFailure)
	totalSCountCounter = append(totalSCountCounter, g.FalconCMDBHostNameAlarmSuccess)
	totalSCountCounter = append(totalSCountCounter, g.FalconPingSuccess)
	totalSCountCounter = append(totalSCountCounter, g.FalconPingFailuer)
	totalSCountCounter = append(totalSCountCounter, g.FalconPingAlarmSuccess)
	totalSCountCounter = append(totalSCountCounter, g.FalconPingCheckSuccess)
	totalSCountCounter = append(totalSCountCounter, g.FalconPingCheckFailuer)

	for _, scount := range totalSCountCounter {
		metricType := "COUNTER"
		metric := transferMetrics(scount, metricType)
		sendMetricsCounter = append(sendMetricsCounter, metric)
	}

	sendMetrics = append(sendMetricsGauge, sendMetricsCounter...)

	return

}


func transferMetrics(metric *g.SCount, metricType string) (falconMetrics MetricValue){

	step := g.Config().FalconAgent.Step
	uts := time.Now().Unix()

	falconMetrics.Endpoint = HOSTNAME
	falconMetrics.Metric = metric.Name
	falconMetrics.Value = metric.Cnt
	falconMetrics.Type = metricType
	falconMetrics.Timestamp = uts
	falconMetrics.Step = step

	return falconMetrics
}
