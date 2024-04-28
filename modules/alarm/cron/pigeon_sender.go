//add by vincent.zhang for pigeon
package cron

import (
	"errors"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
	"github.com/toolkits/net/httplib"
)

var (
	pigeonApp = model.PigeonApp{Source: PIGEON_SOURE_NAME, Key: PIGEON_KEY}
)

func falconLevelToPigeonLevel(level int) string {
	var pigeon_priority string
	if level == 0 {
		pigeon_priority = "5"
	} else if level == 1 {
		pigeon_priority = "4"
	} else if level == 2 {
		pigeon_priority = "3"
	} else {
		pigeon_priority = "2"
	}
	return pigeon_priority
}

func createPigeonAlarm(pigeon *model.Pigeon) (string, error) {
	if pigeon.Status != "PROBLEM" {
		return "", errors.New("createPigeonSend don't support status:" + pigeon.Status)
	}
	priority := falconLevelToPigeonLevel(pigeon.Priority)
	fid_string := strconv.FormatInt(pigeon.Fid, 10)
	alarm := model.PigeonAlarm{
		Fid:       fid_string,
		AlarmCode: pigeon.AlarmCode,
		Value:     pigeon.Value,
		Subject:   pigeon.Subject,
		Sms:       pigeon.Sms,
		Message:   pigeon.Message,
		Priority:  priority,
		Host:      pigeon.Host,
		HostName:  pigeon.HostName,
		Domain:    "",
		Transfer:  pigeon.Transfer,
		AlarmTime: pigeon.AlarmTime,
		ExtArgs:   pigeon.ExtArgs,
	}

	element := model.PigeonAlarmsElement{
		Alarms: []*model.PigeonAlarm{&alarm},
		App:    &pigeonApp,
	}

	data := model.PigeonAlarmsData{
		PigeonElem: &element,
	}

	send := model.PigeonAlarmsSend{
		Data:        &data,
		RequestType: "json",
	}

	return send.String()
}

func createPigeonOK(pigeon *model.Pigeon) (string, error) {
	if pigeon.Status != "OK" {
		return "", errors.New("createPigeonOK don't support status:" + pigeon.Status)
	}
	if pigeon.Fid == PIGEON_DEFAULT_FID {
		return "", errors.New("get default fid for ok event.")
	}
	fid_string := strconv.FormatInt(pigeon.Fid, 10)
	//处理没有ip信息，使用hostname
	host := pigeon.Host
	if host == "" {
		host = pigeon.HostName
	}
	alarm := model.PigeonOK{
		Fid:    fid_string,
		Status: "2",
		//modified by vincent.zhang for  pigeon cleaning with hosgroup is not in cmdb
		//Domain:  pigeon.Domain,
		Domain:  "",
		Host:    host,
		EndTime: pigeon.AlarmTime,
	}

	element := model.PigeonOKElement{
		Alarms: []*model.PigeonOK{&alarm},
		App:    &pigeonApp,
	}

	data := model.PigeonOKData{
		PigeonElem: &element,
	}
	return data.String()
}

func ConsumePigeon() {
	for {
		L := redi.PopAllPigeon()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendPigeonList(L)
	}
}

func SendPigeonList(L []*model.Pigeon) {
	for _, pigeon := range L {
		PigeonWorkerChan <- 1
		//add by vincent.zhang, get fid need to serial call
		if pigeon == nil {
			continue
		}
		// add end
		go SendPigeon(pigeon)
	}
}

func SendPigeon(pigeon *model.Pigeon) {
	defer func() {
		<-PigeonWorkerChan
	}()
	if pigeon.Status == "PROBLEM" {
		url := g.Config().Pigeon.AlarmAddr
		sendPigeonAlarm(url, pigeon)
	} else if pigeon.Status == "OK" {
		url := g.Config().Pigeon.OKAddr
		sendPigeonOK(url, pigeon)
	} else {
		return
	}
}

func sendPigeonAlarm(url string, pigeon *model.Pigeon) {
	requestData, err := createPigeonAlarm(pigeon)
	if err != nil {
		log.Errorf("create pigeon alarm fail, error:%s", err.Error())
		return
	}
	req := httplib.Post(url).SetTimeout(5*time.Second, 30*time.Second)
	req.Body(requestData)
	resp := PigeonResponse{}
	err = req.ToJson(&resp)
	if resp.Object != "" {
		pigeonReturnFid := resp.Object
		log.Infof("send pigeon alarm resp:%v, pigeonid: %v, request:%s, url:%s", resp, pigeonReturnFid, requestData, url)
	} else {
		log.Infof("send pigeon alarm resp:%v, request:%s, url:%s", resp, requestData, url)
	}
	if err != nil {
		log.Errorf("send pigeon alarm fail, error:%s", err.Error())
		return
	}
	if resp.Success == false {
		log.Errorf("send pigeon alarm fail, message:%s", resp.Message)
	}
	return
}

func sendPigeonOK(url string, pigeon *model.Pigeon) {
	requestData, err := createPigeonOK(pigeon)
	if err != nil {
		log.Errorf("create pigeon ok fail, error:%s", err.Error())
		return
	}
	req := httplib.Get(url).SetTimeout(5*time.Second, 30*time.Second)
	req.Param("data", requestData)
	req.Param("requestType", "json")

	resp := PigeonResponse{}
	err = req.ToJson(&resp)

	log.Infof("send pigeon ok resp:%v, request:%s, url:%s", resp, requestData, url)
	if err != nil {
		log.Errorf("send pigeon ok fail, error:%s", err.Error())
		return
	}
	if resp.Success == false {
		log.Errorf("send pigeon ok fail, message:%s", resp.Message)
	}
	return
}
