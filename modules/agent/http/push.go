package http

import (
	"encoding/json"
	"github.com/signmem/falcon-plus/common/model"
	cutils "github.com/signmem/falcon-plus/common/utils"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"net/http"
	"regexp"
	"strings"
)

func configPushRoutes() {
	http.HandleFunc("/v1/push", func(w http.ResponseWriter, req *http.Request) {
		if req.ContentLength == 0 {
			http.Error(w, "body is blank", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(req.Body)
		var metrics []*model.MetricValue
		err := decoder.Decode(&metrics)
		if err != nil {
			http.Error(w, "connot decode body", http.StatusBadRequest)
			return
		}
		standardMetric := formatMetric(metrics)
		g.SendToTransfer(standardMetric)
		w.Write([]byte("success"))
	})
}

func formatMetric(uploadMetric  []*model.MetricValue ) ( formatMetric []*model.MetricValue) {
	IllegalChar := g.Config().IllegalChar
	IllegalCharString := strings.Join(IllegalChar,"|")

	var re = regexp.MustCompile(IllegalCharString)
	for _, v := range uploadMetric {
		var staticMetric *model.MetricValue
		staticMetric = new(model.MetricValue)
		staticMetric.Metric =  re.ReplaceAllString(v.Metric, "")

		if len(v.Tags) > 0 {
			var validTags []string
			tags := strings.Split(v.Tags, ",")

			for _, tag := range tags {
				if cutils.IsChinese(tag) {
					continue
				}

				splitTag := strings.Split(tag, "=")
				if len(splitTag) != 2 {
					continue
				}

				if  len(splitTag[1]) == 0 {
					continue
				}

				validTags = append(validTags,  re.ReplaceAllString(tag,""))
			}

			if len(validTags) > 0 {
				staticMetric.Tags = strings.Join(validTags,",")
			} else {
				staticMetric.Tags = ""
			}

		}
		staticMetric.Value = v.Value
		staticMetric.Type  = v.Type
		staticMetric.Endpoint = v.Endpoint
		staticMetric.Step  = v.Step
		staticMetric.Timestamp = v.Timestamp
		formatMetric = append(formatMetric, staticMetric)
	}
	return formatMetric
}