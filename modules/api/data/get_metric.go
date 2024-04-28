package data

import (
	"bufio"
	"io/ioutil"
	"github.com/open-falcon/falcon-plus/modules/api/config"
	"github.com/open-falcon/falcon-plus/common/http"
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	"encoding/json"
	"os"
	"time"
)

var (
	MetricFile string
	MetricSliceNow []string
)

func CronGenMetric() {
	time.Sleep(time.Second * time.Duration(10))
	var plan string
	for {
		config.Logger.Debug("CronGenMetric() get metrice start.")

		if len(MetricSliceNow) == 0 {
			plan = "full"

		}else {
			plan = "append"
		}

		// need do someting.
		metricDataFlush := getMetricData(plan)

		var appendMetric []string

		for _, metric := range metricDataFlush {
			if cutils.StringInSlice(metric, MetricSliceNow) == false {
				MetricSliceNow = append(MetricSliceNow, metric)
				appendMetric = append(appendMetric, metric)
			}
		}

		// MetricSliceNow = metricData

		// 避免了过节期间没有主机更新导致犯规 metricData 为 0 或丢失旧期历史数据

		err := AppendMetric(appendMetric)
		if err != nil {
			config.Logger.Errorf("CronGenMetric() error with: %s", err)
		}

		time.Sleep(time.Second * time.Duration(3600))
	}
}

func AppendMetric(metrics []string) (err error) {
	f, err := os.OpenFile(MetricFile,  os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, i := range metrics {
		if _, err = f.WriteString(i + "\n"); err != nil {
			return err
		}
	}
	return nil
}


func CountMetricFileLine() int {
	fileName := MetricFile
	file,err := os.Open(fileName)
	count := 0
	if err != nil{
		return count
	}
	defer file.Close()
	fd:=bufio.NewReader(file)

	for {
		_,err := fd.ReadString('\n')
		if err!= nil{
			break
		}
		count++
	}
	return count
}

func getMetricData(plan string) ([]string) {
	var data []string
	params := "?plan=" + plan
	falconDeployTypeApi := "http://127.0.0.1:8080/api/v1/graph/endpoint_counter_distinct"
	resp, err := http.HttpApiGet(falconDeployTypeApi, params, "falcon")

	if err != nil {
		config.Logger.Errorf("getMetricData() httpapi get metric data error.")
		return data
	}

	responseBody, err := ioutil.ReadAll(resp)
	if err != nil {
		config.Logger.Errorf("getMetricData() error with %s", err)
	}
	defer resp.Close()

	_ = json.Unmarshal(responseBody, &data)
	return data
}


func writefile( charString  []string) bool {
	file, err := os.OpenFile(MetricFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		config.Logger.Errorf  ("failed creating file: %s", err)
		return false
	}

	datawriter := bufio.NewWriter(file)

	for _, data := range charString {
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	file.Close()
	return true
}
