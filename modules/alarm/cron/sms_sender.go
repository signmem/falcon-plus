package cron

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/signmem/falcon-plus/modules/alarm/api"
	"github.com/signmem/falcon-plus/modules/alarm/g"
	"github.com/signmem/falcon-plus/modules/alarm/model"
	"github.com/signmem/falcon-plus/modules/alarm/redi"
)

var transfer, counter string

func ConsumeSms() {
	for {
		L := redi.PopAllSms()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendSmsList(L)
	}
}

func SendSmsList(L []*model.Sms) {
	for _, sms := range L {
		SmsWorkerChan <- 1
		go SendSms(sms)
	}
}

func SendSms(sms *model.Sms) {
	defer func() {
		<-SmsWorkerChan
	}()

	query := make(map[string]interface{})

	//data层
	data := make(map[string]interface{})

	//alarms层
	alarms := make([]map[string]string, 0, 10)

	//app层
	app := make(map[string]string)

	//赋值
	subject := sms.Content

	//赋值alarm_time
	fi := strings.LastIndex(subject, "[")
	si := strings.LastIndex(subject, "]")
	alarm_time := string(subject[fi+1 : si])

	//赋值pigeon_priority
	f1 := strings.Index(subject, "[")
	s1 := strings.Index(subject, "]")
	priority := subject[f1+1 : s1]
	priority = string(priority)
	var pigeon_priority string
	if priority == "L0" {
		pigeon_priority = "5"
	} else if priority == "L1" {
		pigeon_priority = "4"
	} else if priority == "L2" {
		pigeon_priority = "3"
	} else {
		pigeon_priority = "2"
	}

	//赋值host
	h := strings.NewReplacer("[", " ", "]", "")
	subject = h.Replace(subject)
	subject_fields := strings.SplitN(subject, " ", -1)
	host := subject_fields[3]
	/*
		hostinfo1 := g.GetHostIP(host)
		hostip := string(hostinfo1.IP)
		hostinfo2 := g.GetHostGroup(host)
		hostgroup := hostinfo2.Hostgroup
	*/
	hostip, _ := g.GetHostIP(host)
	hostgroup, _ := g.GetHostGroup(host)

	//赋值message2pigeon
	message := strings.Join(subject_fields[1:len(subject_fields)-2], " ")
	message2pigeon := message + " (" + hostip + ")"

	//赋值subject2pigeon
	subject2pigeon := "<" + hostgroup + "> " + hostip + " " + strings.Join(strings.Fields(message)[3:], " ")

	//赋值transfer
	s3 := strings.Split(message, " ")
	metric := s3[4]
	tag := s3[5]
	if len(tag) == 0 {
		counter = metric
	} else {
		counter = metric + "/" + tag
	}
	input := &api.APITmpGraphInput{
		Endpoints: []string{host},
		Counters:  []string{counter},
	}

	ret := api.CreateTmpGraph(input)
	log.Debugf("endpoint is %s", host)
	log.Debugf("counter is %s", counter)
	if ret != nil {
		log.Debugf("id is %d", ret.Id)
		alarm_id := strconv.Itoa(ret.Id)
		transfer = "http://falcon-dashboard.vip.vip.com/charts?id=" + alarm_id + "&graph_type=k"
		log.Debugf("transfer is %s", transfer)
	}

	alarms = append(alarms, map[string]string{
		"fid":        "9560",
		"value":      "100",
		"subject":    subject2pigeon,
		"message":    message2pigeon,
		"priority":   pigeon_priority,
		"host":       hostip,
		"domain":     hostgroup,
		"transfer":   transfer,
		"alarm_time": alarm_time,
	})

	//赋值
	app["source"] = "vip-falcon"
	app["key"] = "99a35ace8952f662cad672d2f6da5754"

	//赋值
	data["pigeon"] = map[string]interface{}{
		"alarms": alarms,
		"app":    app,
	}

	//把请求的结构转化为请求规范的json字符串
	data_json, _ := json.Marshal(data)
	query["data"] = string(data_json)
	query["requestType"] = "json"
	data_query, _ := json.Marshal(query)
	buf := bytes.NewBuffer(data_query)

	url := g.Config().Api.Sms
	resp, err := http.Post(url, "application/json", buf)

	if err != nil {
		log.Println(err)
	}

	log.Debugf("send sms:%v, resp:%v, url:%s", sms, resp, url)
}
