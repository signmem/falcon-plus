package pigeon

import (
	"fmt"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"strings"
	"time"
	"encoding/json"
)

func GenFalconView(alarminfo Alarm) string {
	metric := alarminfo.Metric
	hostname := alarminfo.Hostname
	newchar := strings.Replace(metric, ".", "_", -1)

	return fmt.Sprintf("https://m3.vip.vip.com/v3/dashboard/panel/view/255/0?refresh=1m&orgId=1&from=now-3h&to=now&var-host_name=%s&var-metric=%s", hostname,metric)
}

func GenAlarmSubject (alarminfo Alarm) string {
	return fmt.Sprintf("[%s] %s (%s) %s", alarminfo.Domain, alarminfo.Hostname,
		alarminfo.Ip, alarminfo.Event )
}

func GenAlarmMesage (alarminfo Alarm) string {
	return fmt.Sprintf("[%s] %s (%s) %s, 当前值：%s", alarminfo.Domain,
		alarminfo.Hostname,	alarminfo.Ip, alarminfo.Detail , alarminfo.Value)
}

func GenSmsSubject(alarminfo Alarm) string {
	// sms  informastion == message information !
	return fmt.Sprintf("[%s] %s (%s) %s, 当前值：%s", alarminfo.Domain, alarminfo.Hostname,
		alarminfo.Ip, alarminfo.Event, alarminfo.Value)
}

func buildPigeonExtagrLegend(alarminfo Alarm) string {
	return fmt.Sprintf("[%s] %s", alarminfo.Domain, alarminfo.Hostname,)
}

func buildPigeonExtagrTitil(alarminfo Alarm) string {
	titleInfo := fmt.Sprintf("%s",alarminfo.Event)
	return strings.Replace(titleInfo,"%","",-1)
}

func buildPigeonM3Chart(alarminfo Alarm) M3Chart {
	var newM3Chart M3Chart
	newM3Chart.Legend = buildPigeonExtagrLegend(alarminfo)
	newM3Chart.Title =  buildPigeonExtagrTitil(alarminfo)
	return newM3Chart
}

func buildPigeonM3ExtraArgs(alarminfo Alarm) *ExtArg {
	var newM3Value M3Value
	var newM3ValueSlice []M3Value
	newM3Value.Body = buildPigeonM3Body(alarminfo)
	newM3Value.Chart = buildPigeonM3Chart(alarminfo)
	newM3Value.URL = g.Config().Pigeon.M3dbUrl
	newM3ValueSlice = append(newM3ValueSlice, newM3Value )
	m3ValueJson, _ := json.Marshal(newM3ValueSlice)
	m3Values := fmt.Sprintf(string(m3ValueJson))
	m3ExtraArgs := ExtArg{
		Name : "wx_chart_url",
		Value: m3Values,
	}
	return &m3ExtraArgs
}

func buildPigeonExtagrMetric(alarminfo Alarm) string {
	m3dbMetricName :=  fmt.Sprintf(
		"%s",
		strings.Replace(alarminfo.Metric,".","_",-1),
	)

	hostinfo := fmt.Sprintf("host_name=\"%s\"", alarminfo.Hostname)
	m3dbSearchMetric := fmt.Sprintf("%s{%s}", m3dbMetricName, hostinfo)
	return m3dbSearchMetric
}

func buildPigeonM3Body(alarminfo Alarm) string {

	var newM3Body M3Body

	metricInfo := buildPigeonExtagrMetric(alarminfo)
	newM3Body.Metric      = fmt.Sprintf("%s", metricInfo)
	newM3Body.DatasourceID = 3
	newM3Body.From         = time.Now().Unix() - 3600
	newM3Body.Step         = 60
	newM3Body.Source       = "81a7b791e01f4b899dc2ceaa047b25d7"
	newM3Body.To           = time.Now().Unix()
	newM3Body.Type         = "query_range"
	newM3BodyJson, _ := json.Marshal(newM3Body)
	return fmt.Sprintf(string(newM3BodyJson))
}





