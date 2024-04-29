//add by vincent.zhang for pigeon
package model

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type PigeonApp struct {
	Source string `json:"source"`
	Key    string `json:"key"`
}

type ExtArg struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type M3Body struct {
        Metric       string `json:"metric"`
        DatasourceID int    `json:"datasourceId"`
        From         int64  `json:"from"`
        Step         int    `json:"step"`
        Source       string `json:"source"`
        To           int64  `json:"to"`
        Type         string `json:"type"`
}

type M3Chart struct {
        Legend string `json:"legend"`
        Title  string `json:"title"`
}

type M3Value struct {
        Body    string `json:"body"`
        Chart   M3Chart `json:"chart"`
        URL string `json:"url"`
}

//alarm event
type PigeonAlarm struct {
	Fid       string    `json:"fid"`
	AlarmCode string    `json:"alarm_code"`
	Value     string    `json:"value"`
	Subject   string    `json:"subject"`
	Sms       string    `json:"sms"`
	Message   string    `json:"message"`
	Priority  string    `json:"priority"`
	Host      string    `json:"host"`
	HostName  string    `json:"hostname"`
	Domain    string    `json:"domain"`
	Transfer  string    `json:"transfer"`
	AlarmTime string    `json:"alarm_time"`
	ExtArgs   []*ExtArg `json:"ext_args"`
}

type PigeonAlarmsElement struct {
	Alarms []*PigeonAlarm `json:"alarms"`
	App    *PigeonApp     `json:"app"`
}

type PigeonAlarmsData struct {
	PigeonElem *PigeonAlarmsElement `json:"pigeon"`
}

type PigeonAlarmsSend struct {
	Data        *PigeonAlarmsData `json:"data"`
	RequestType string            `json:"requestType"`
}

func (this *PigeonAlarmsSend) String() (string, error) {
	data_query, err := json.Marshal(this)
	if err != nil {
		log.Errorf("json marshal PigeonAlarmsSend fail error:%s, object:%v", err.Error(), this)
		return "", err
	}
	return string(data_query), nil
}

// OK event
type PigeonOK struct {
	Fid     string `json:"fid"`
	Status  string `json:"status"`
	Domain  string `json:"domain"`
	Host    string `json:"host"`
	EndTime string `json:"end_time"`
}

type PigeonOKElement struct {
	Alarms []*PigeonOK `json:"alarms"`
	App    *PigeonApp  `json:"app"`
}

type PigeonOKData struct {
	PigeonElem *PigeonOKElement `json:"pigeon"`
}

func (this *PigeonOKData) String() (string, error) {
	data_query, err := json.Marshal(this)
	if err != nil {
		log.Errorf("json marshal PigeonOKData fail error:%s, object:%v", err.Error(), this)
		return "", err
	}
	return string(data_query), nil
}

// Event in redis, prepare to send to pigeon
type Pigeon struct {
	Fid       int64     `json:"fid"`
	AlarmCode string    `json:"alarm_code"`
	Status    string    `json:"status"` // OK or PROBLEM
	Value     string    `json:"value"`
	Subject   string    `json:"subject"`
	Sms       string    `json:"sms"`
	Message   string    `json:"message"`
	Priority  int       `json:"priority"`
	Host      string    `json:"host"`
	HostName  string    `json:"hostname"`
	Domain    string    `json:"domain"`
	Transfer  string    `json:"transfer"`
	AlarmTime string    `json:"alarm_time"`
	ExtArgs   []*ExtArg `json:"ext_args"`
}

func (this *Pigeon) String() (string, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		log.Errorf("json marshal Pigeon fail error:%s, object:%v", err.Error(), this)
		return "", err
	}
	return string(bs), err
}
