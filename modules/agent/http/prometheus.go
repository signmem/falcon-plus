package http

import (
	"encoding/json"
	"errors"
	fhttp "github.com/signmem/falcon-plus/common/http"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/prometheus/common/expfmt"
	dto "github.com/prometheus/client_model/go"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
	"fmt"
)


var (
        STEP  				int64
        ValidMetric 		[]string  // 用于过滤并只收集当前 list 中的 metrics 其他 metrics 不需要上报
        MatchCalMetricList 	[]string  // 用于匹配自身计算用的 metrics, 匹配后，更新 CalMetric 中的值
        CalMetricDict  		[]g.MetricCalType // 用于自身 metrics 计算用的常量，被动更新
        SumMetrics			[]string  // use to sum metrics values
        AllMetrics			[]string  // use to match metrics
        floatType = reflect.TypeOf(float64(0))
)


func configPrometheusRoutes() {

	http.HandleFunc("/upload/prometheus/metrics",
		func(w http.ResponseWriter, r *http.Request){

		if r.Method != http.MethodPost {
			http.Error(w, "POST method support only.", http.StatusMethodNotAllowed)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "json format support only.", http.StatusUnsupportedMediaType)
			return
		}

		var promJson g.PromConfig
		decoder := json.NewDecoder(r.Body)
		// decoder.DisallowUnknownFields()

		if err := decoder.Decode(&promJson); err != nil {
			log.Printf("[ERROR] configPrometheusRoutes() json decode err: %s", err)
			http.Error(w, "json format not valid: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if promJson.SslEnable == true {
			if promJson.TLS.KeyFile == "" ||
				promJson.TLS.CertFile == "" ||
				promJson.TLS.CaFile == "" {

				http.Error(w, "SSL json value not valid", http.StatusBadRequest)
				return
			}

			if fileExists(promJson.TLS.KeyFile) == false ||
				fileExists(promJson.TLS.CertFile) == false ||
				fileExists(promJson.TLS.CaFile) == false {

				http.Error(w, "SSL file not exists", http.StatusBadRequest)
				return
			}
		}

		STEP = promJson.Step
		valiidMetricFile	:= promJson.ValidMetricFile
		sumMetricFile 		:= promJson.SumMetricFile
		calMetricFile		:= promJson.CalMetricFile

		ValidMetric, _	= LoadMetricJsonFile(valiidMetricFile)
		SumMetrics, _ 	= LoadMetricJsonFile(sumMetricFile)
		MatchCalMetricList, CalMetricDict, _ =	LoadCalMetricJsonFile(calMetricFile)

		AllMetrics = append(append(ValidMetric, MatchCalMetricList...),SumMetrics...)

		getMetrics(promJson)

		w.Write([]byte("prometheus send ok."))

	})
}



func getMetrics(promConfig g.PromConfig) (getAllMetric []*model.MetricValue) {
        // 用于获取 对应 prometheus 服务器 /metrics 信息
        // 通过 /metrics 中 TYPE 中信息定义 type, 包含 (counter gauge histogram summary) || untyped

	metricsString, err := getMetricFromServer(promConfig)  // 调用 http 获取指定服务器 metrics

	if err != nil {
		log.Printf("[ERROR] getMetrics() call getMetricFromServer() error: %s\n", err)
		return
	}

	if len(metricsString) == 0 && promConfig.Debug == true  {
		log.Println("[DEBUG] getMetrics() call getMetricFromServer() get zero metrics.")
		return
	}

	getAllMetric, matchMetricDict, err := genMetricFormat(metricsString, promConfig.Debug)

	if err != nil {
		log.Printf("[ERROR] getMetrics() call genMetricFormat() error: %s\n", err)
		return
	}

	if  len(MatchCalMetricList) > 0 && len(CalMetricDict) > 0 {
		calMatchMetrics := calMatchMetricValues(CalMetricDict, matchMetricDict)
		getAllMetric = append(getAllMetric, calMatchMetrics...)
	}

	if promConfig.Debug == true {
		log.Printf("[DEBUG]: getMetrics() total metrics: %d\n", len(getAllMetric))
	}

	var resp model.TransferResponse

	g.SendMetrics(getAllMetric, &resp)

	return
}

func genMetricFormat(info string, debug bool) (getAllMetric []*model.MetricValue,
	matchMetricDict []*model.MetricValue, err error) {


	// getAllMetric  all of the metrics (包含了计算前的 metrics)
	// matchMetricDict 只包含用于计算用的 metrics 信息

	timestamp := time.Now().Unix()

	var metric *model.MetricValue

	// totalCountMetric  use to auto sum values
	var totalCountMetric []*model.MetricValue

	parser := &expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(strings.NewReader(info))

	if err != nil {
		log.Printf("[ERROR]: failed to parse input: %s\n", err)
		return getAllMetric, matchMetricDict,  err
	}

	for _, val := range families {

		for _, m := range val.GetMetric() {

			metric = &model.MetricValue{}
			metric.Metric = val.GetName()

			// 只有自定义需要获取的 metrics 时候才需要执行下面过滤操作
			// 如果不过滤，则获取所有 metrics
			if (len(AllMetrics)) > 0 {
				if KeyinSliceWithChar(AllMetrics, metric.Metric) == false {
					continue
				}
			}

			// 或者 mertric 对应 values, 根据不同类型进行判定
			switch val.GetType() {
			case dto.MetricType_COUNTER:
				metric.Value = m.GetCounter().GetValue()
				metric.Type = "COUNTER"
			case dto.MetricType_GAUGE:
				metric.Value = m.GetGauge().GetValue()
				metric.Type = "GAUGE"
			case dto.MetricType_UNTYPED:
				metric.Value = m.GetUntyped().GetValue()
				metric.Type = "GAUGE"
			case dto.MetricType_SUMMARY:
				metric.Value = m.GetSummary().GetSampleSum()
				metric.Type = "SUMMARY"
			default:
				// 部分 metrics type unkonw , 暂时作为 gauge type
				metric.Value = m.GetGauge().GetValue()
				metric.Type = "GAUGE"
			}

			metric.Tags = ""
			var metricTags string

			metric.Metric = val.GetName()
			metric.Step = STEP
			metric.Endpoint = g.HostName
			metric.Timestamp = timestamp

			// 获取指标的 metrics

			for n, label := range m.GetLabel() {
				if n == len(m.Label) - 1 {
					tag := fmt.Sprintf("%s=%s" , label.GetName(), label.GetValue())
					metric.Tags = metricTags + tag
				} else {
					// 多个 tags 处理方法
					tag := fmt.Sprintf("%s=%s," , label.GetName(), label.GetValue())
					metric.Tags = metricTags + tag
				}
			}

			// 对需要计算的 metric 进行收集， 放外部进行处理 (matchMetricDict)
			if (len(MatchCalMetricList) > 0 ){
				for _, name := range MatchCalMetricList {
					if name == metric.Metric {
						matchMetricDict = append(matchMetricDict, metric)
					}
				}
			}

			// 需要对计算 count 总数的 metric 进行处理， 避免过多 metric 进行上报
			var appendGrap bool
			if (len(SumMetrics) > 0) {
				for _, name := range SumMetrics {
					if metric.Metric == name {
						totalCountMetric = metricAddSlice(metric, totalCountMetric)
						// not going to send
						appendGrap = true
						break
					}
				}
			}

			if appendGrap == true {
				continue
			}

			getAllMetric = append(getAllMetric, metric)
		}
	}

	if debug == true {
		log.Printf("[DEBUG] totalCountMetric len:%d\n", len(totalCountMetric))
	}

	// count metrics 只需要在循环结束后放入 getAllMetric 变量中
	getAllMetric = append(getAllMetric, totalCountMetric...)

	if g.Config().Debug || debug == true {
		for _, metric := range getAllMetric {
			log.Printf("[PROMETHEUS] metrics: %s\n", metric.String())
		}
	}

	return getAllMetric, matchMetricDict,nil

}



func getMetricFromServer(promConfig g.PromConfig) (httpresponse string, err error) {

	var metricFromHTTP []byte = []byte{}

	server := promConfig.ServerConfig.Server
	port := promConfig.ServerConfig.Port
	metricAPI := promConfig.ServerConfig.MetricAPI

	if promConfig.SslEnable == false {
		url := "http://" + server + ":" + port + metricAPI
		metricFromHTTP, err = fhttp.HttpApiGet(url, "","")
		if err != nil {
			log.Printf("[ERROR] GetMetricFromPrometheus() error:%s\n", err)
			return "", err
		}
	}

	if promConfig.SslEnable == true {

		var sslfile fhttp.SSLCert
		sslfile.CaFile   = promConfig.TLS.CaFile
		sslfile.CertFile = promConfig.TLS.CertFile
		sslfile.KeyFile  = promConfig.TLS.KeyFile

		url := "https://" + server + ":" + port + metricAPI
		metricFromHTTP, err = fhttp.HttpsApiGet(url, "", sslfile)
		if err != nil {
			log.Printf("[ERROR] GetMetricFromPrometheus() HttpsApiGet() error:%s\n", err)
			return "", err
		}

	}

	info := string(metricFromHTTP)
	lineCount := strings.Count(info, "\n")

	if promConfig.Debug == true {
		log.Printf("[DEBUG] GetMetricFromPrometheus() HttpsApiGet() return %d lines\n", lineCount)
	}

	return info, nil
}


func calMatchMetricValues(calMetricDict []g.MetricCalType,
        calMetric []*model.MetricValue) (calMatchMetric []*model.MetricValue) {


	// calMetricList = list == prom.MatchCalMetricList
	// calMetricDict = dict == prom.CalMetricDict
	// calMetric = 只包含 prom.MatchCalMetricList 中的所有 metrics 信息
	// calMatchMetric = 通过计算 metricsum / metriccount 对应的 metricname

	for _, calInfo := range calMetricDict {

		sumName := calInfo.MetricSum
		var tags []string

		for _, metricInfo := range  calMetric {

			if metricInfo.Metric == sumName {

				if KeyinSliceWithChar(tags, metricInfo.Tags) == false {
					tags = append(tags, metricInfo.Tags)
				}
			}
		}

		for _, tag := range tags {

			var newMetric *model.MetricValue
			var countName, metricName string
			var countVale, sumValue interface{}

			for _, info := range calMetricDict {

				if info.MetricSum == sumName {
					countName = info.MetricCount
					metricName = info.MetricName
					break
				}
			}

			newMetric.Metric = metricName

			for _, metricInfo := range calMetric {
				if metricInfo.Metric == sumName && metricInfo.Tags == tag {
					sumValue = metricInfo.Value
				}

				if metricInfo.Metric == countName && metricInfo.Tags == tag {
					countVale = metricInfo.Value
					newMetric.Endpoint = metricInfo.Endpoint
					newMetric.Timestamp = metricInfo.Timestamp
					newMetric.Step = metricInfo.Step
					newMetric.Type = metricInfo.Type
				}
			}

			if GetValueToFloat(countVale) == 0 {
				newMetric.Value = float64(0)

			} else {
				newMetric.Value = GetValueToFloat(sumValue) / GetValueToFloat(countVale)
			}

			newMetric.Tags = tag

			calMatchMetric = append(calMatchMetric, newMetric)
		}

	}

	return calMatchMetric

}

func metricAddSlice(newStruct *model.MetricValue, w []*model.MetricValue)(q []*model.MetricValue) {

	// 用于自动化把 metricValue 加入 []MetricValue
	// 自动化对 value 进行 sum 计算

	if metricInSlice(newStruct, w) == true {
		q = metricAddValue(newStruct, w)
	} else {
		q = append(w, newStruct)
	}
	return q
}


func KeyinSliceWithChar(s []string, j string) (exit bool) {

	for _, i := range s {
		if i == j {
			return true
		}
	}
	return false
}

func GetValueToFloat(unk interface{}) (float64) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0
	}
	fv := v.Convert(floatType)
	return fv.Float()
}

func metricInSlice(newStruct *model.MetricValue, w []*model.MetricValue) bool {
	// use to judge metric in []metrics

	for _, info := range w {
		if info.Metric == newStruct.Metric {
			return true
		}
	}
	return false
}


func metricAddValue(newStruct *model.MetricValue, w []*model.MetricValue)(q []*model.MetricValue) {

	for _, metric := range w {
		if metric.Metric == newStruct.Metric {
			metric.Add(newStruct.Value)
		}
		q = append(q, metric)
	}
	return
}



func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func LoadMetricJsonFile (filePath string) (validMetric []string, err error) {

	if _, err := os.Stat(filePath); err == nil {

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("[ERROR] loadjsonfile %s read error\n", filePath)
			return validMetric, err
		}

		err = json.Unmarshal(content, &validMetric)

		if err != nil {
			log.Printf("[ERROR] loadjsonfile %s json format error\n", filePath)
			return validMetric, err
		}

	} else {
		log.Printf("[ERROR] loadjsonfile %s not exists\n", filePath)
		return validMetric, err
	}
	return validMetric, nil
}


func LoadCalMetricJsonFile(filePath string) (matchMetric []string,
	calMetric []g.MetricCalType, err error) {

	if _, err := os.Stat(filePath); err == nil {

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("[ERROR] cal metrics jsonfile %s read error\n", filePath)
			return matchMetric, calMetric, err
		}

		err = json.Unmarshal(content, &calMetric)

		if err != nil {
			log.Printf("cal metrics jsonfile %s json format error\n", filePath)
			return matchMetric, calMetric, err
		}

		for _, metrics := range calMetric {
			matchMetric = append(matchMetric, metrics.MetricSum)
			matchMetric = append(matchMetric, metrics.MetricCount)
		}

	}
	// log.Printf("metrics jsonfile %s not exists\n", filePath)

	return matchMetric, calMetric, nil
}

