package pigeon

import (
	"encoding/json"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/tools"
	"io/ioutil"
	"time"
)

func GenPigeonAlarmData(alarminfo Alarm) Alarms {
	// use to build metric == agent.alive (default)
	// FID == 14237 (default)
	// AlarmCode == 122-000 (default)
	// Status == "PROBLEM" (DEFAULT)
	// Value ==  var f float64 = -1
	// subject == [domain] hostname (ip) event
	// Sms  == hostname xxxx, value: 0
	// Message == Sms
	// priority = 2
	// host = IP
	// hostname = hostname
	// domain = domain
	// transfer == falcon-view
	// alarm_time = now
	// ext_args = [xxxx]

	// event := "falcon agent异常"
	// metric := "agent.alive"

	subject := GenAlarmSubject(alarminfo)
	smsSubject := GenSmsSubject(alarminfo)
	transfer := GenFalconView(alarminfo)

	timeStr := time.Now().Format("2006-01-02 15:04:05")
	rightArgs := &ExtArg{ Name : "right_value", Value: alarminfo.Value}

	m3ExtArgs := buildPigeonM3ExtraArgs(alarminfo)   // notice

	var defaultArgs []*ExtArg
	defaultArgs = append(defaultArgs, rightArgs)
	defaultArgs = append(defaultArgs, m3ExtArgs)

	var report Alarms

	report = Alarms {
		Fid : alarminfo.Fid,
		AlarmCode : alarminfo.AlarmCode,
		Value: alarminfo.Value,
		Subject: subject,
		Sms: smsSubject,
		Message: alarminfo.Message,
		Priority: alarminfo.Priority,
		Host: alarminfo.Ip,
		HostName: alarminfo.Hostname,
		Domain: alarminfo.Domain,
		Transfer: transfer,
		AlarmTime: timeStr,
		ExtArgs: defaultArgs,
	}

	return report
}

func LogPigeonAlarm(alarminfo Alarms, pigeonID string ) {
	g.Alarmer.Infof("[pigeon 返回信息] LogPigeonAlarm() id: %s 详细信息: %s",
		pigeonID, alarminfo.Message)
}

func SendPigeonAlarm(p Alarm) (err error){

	pigeonAlarms := GenPigeonAlarmData(p)
	var feedback GenReport
	hostname := pigeonAlarms.HostName
	feedback.Data.Pigeon.Alarm = append(feedback.Data.Pigeon.Alarm, pigeonAlarms)
	feedback.Data.Pigeon.App.Key = g.Config().Pigeon.PigeonKey
	feedback.Data.Pigeon.App.Source = g.Config().Pigeon.PigeonSource

	reportBytes, err := json.Marshal(feedback)

	if err != nil {
		return  err
	}

	pigeonUrl := g.Config().Pigeon.PigeonUrl

	if g.Config().Debug {
		g.Alarmer.Debugf("[发送 pigeon 告警] SendPigeonAlarm() 向 pigeon 发送的告警信息: %v", string(reportBytes))
	}

	resp, err := tools.HttpApiPost(pigeonUrl, reportBytes, "")
	if err != nil {
		g.Logger.Errorf("SendPigeonAlarm() hostname %s acces pigeonurl error %s", hostname, err)
		return err
	}

	respBody, err := ioutil.ReadAll(resp)
	if err != nil {
		g.Logger.Errorf("SendPigeonAlarm() hostname %s http io read error %s", hostname, err)
		return err
	}

	defer resp.Close()
	var pigeonResp PigeonResopose
	err = json.Unmarshal(respBody, &pigeonResp)
	if err != nil {
		g.Logger.Errorf("SendPigeonAlarm() hostname %s http response json unmarshal error %s",
			hostname, err)
		return err
	}

	if pigeonResp.Success != true {
		g.Logger.Errorf("[告警发送失败] - SendPigeonAlarm() hostname %s pigon response false, msg: %s",
			hostname, pigeonResp.Message)
	} else {
		LogPigeonAlarm(pigeonAlarms, pigeonResp.Object)
	}

	return  nil
}
