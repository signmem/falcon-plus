package cron

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"

	cmodel "github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/alarm/api"
	"github.com/signmem/falcon-plus/modules/alarm/g"
	"github.com/signmem/falcon-plus/modules/alarm/redi"
	//add by vincent.zhang for pigeon
	"github.com/signmem/falcon-plus/common/utils"
	eventmodel "github.com/signmem/falcon-plus/modules/alarm/model/event"
	//end
)

func consume(event *eventmodel.EventDetail, isHigh bool) {
	actionId := event.Event.ActionId()
	/*
		if actionId <= 0 {
			return
		}

		action := api.GetAction(actionId)
		if action == nil {
			return
		}

		if action.Callback == 1 {
			HandleCallback(event, action)
		}
	*/
	//modified by vincent.zhang, ignore action for pigeon
	var action *api.Action
	if actionId <= 0 {
		action = nil
	} else {
		action = api.GetAction(actionId)
		if action != nil && action.Callback == 1 {
			HandleCallback(event.Event, action)
		}
	}
	if isHigh {
		consumeHighEvents(event, action)
	} else {
		consumeLowEvents(event, action)
	}
}

// 高优先级的不做报警合并
func consumeHighEvents(event *eventmodel.EventDetail, action *api.Action) {
	/**
	modified by vincent.zhang
	ignoring alarm with ok status based on config
	send alarm based on url and add pigeon
	**/
	cfg := g.Config()
	if cfg.SendOK == false && event.Event.Status == "OK" {
		return
	}
	if cfg.Pigeon.AlarmAddr != "" {
		consumeToPigeon(event)
	}

	if action == nil {
		return
	}
	//end
	if action.Uic == "" {
		return
	}

	phones, mails, ims := api.ParseTeams(action.Uic)

	smsContent := GenerateSmsContent(event.Event)
	mailContent := GenerateMailContent(event.Event)
	imContent := GenerateIMContent(event.Event)

	if cfg.Api.Sms != "" && event.Event.Priority() <= 3 {
		redi.WriteSms(phones, smsContent)
	}
	if cfg.Api.Mail != "" {
		redi.WriteMail(mails, smsContent, mailContent)
	}
	if cfg.Api.IM != "" {
		redi.WriteIM(ims, imContent)
	}
}

// 低优先级的做报警合并
func consumeLowEvents(event *eventmodel.EventDetail, action *api.Action) {
	//add by vincet.zhang
	if action == nil {
		return
	}
	//end
	if action.Uic == "" {
		return
	}
	/**
	modified by vincent.zhang
	ignoring alarm with ok status based on config
	send alarm based on url
	low level events don't send to pigeon
	**/
	cfg := g.Config()
	if cfg.SendOK == false && event.Event.Status == "OK" {
		return
	}
	if cfg.Api.Sms != "" && event.Event.Priority() <= 3 {
		ParseUserSms(event.Event, action)
	}
	if cfg.Api.Mail != "" {
		ParseUserMail(event.Event, action)
	}
	if cfg.Api.IM != "" {
		ParseUserIm(event.Event, action)
	}
	//end
}

func ParseUserSms(event *cmodel.Event, action *api.Action) {
	userMap := api.GetUsers(action.Uic)

	content := GenerateSmsContent(event)
	metric := event.Metric()
	status := event.Status
	priority := event.Priority()

	queue := g.Config().Redis.UserSmsQueue

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for _, user := range userMap {
		dto := SmsDto{
			Priority: priority,
			Metric:   metric,
			Content:  content,
			Phone:    user.Phone,
			Status:   status,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Error("json marshal SmsDto fail:", err)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Error("LPUSH redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}

func ParseUserMail(event *cmodel.Event, action *api.Action) {
	userMap := api.GetUsers(action.Uic)

	metric := event.Metric()
	subject := GenerateSmsContent(event)
	content := GenerateMailContent(event)
	status := event.Status
	priority := event.Priority()

	queue := g.Config().Redis.UserMailQueue

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for _, user := range userMap {
		dto := MailDto{
			Priority: priority,
			Metric:   metric,
			Subject:  subject,
			Content:  content,
			Email:    user.Email,
			Status:   status,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Error("json marshal MailDto fail:", err)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Error("LPUSH redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}

func ParseUserIm(event *cmodel.Event, action *api.Action) {
	userMap := api.GetUsers(action.Uic)

	content := GenerateIMContent(event)
	metric := event.Metric()
	status := event.Status
	priority := event.Priority()

	queue := g.Config().Redis.UserIMQueue

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for _, user := range userMap {
		dto := ImDto{
			Priority: priority,
			Metric:   metric,
			Content:  content,
			IM:       user.IM,
			Status:   status,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Error("json marshal ImDto fail:", err)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Error("LPUSH redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}

//add by vincent.zhang for pigeon
func isNeedPigeonHighCombine(priority int) bool {
	cfg := g.Config()
	for _, p := range cfg.Pigeon.HighCombiner.Levels {
		if priority == p {
			return true
		}
	}
	return false
}

func isNeedPigeonLowCombine(priority int) bool {
	cfg := g.Config()
	for _, p := range cfg.Pigeon.LowCombiner.Levels {
		if priority == p {
			return true
		}
	}
	return false
}

func writePigeonDto(event *cmodel.Event, ip, hostgroup, queue string) {
	if event == nil || queue == "" {
		return
	}

	rc := g.RedisConnPool.Get()
	defer rc.Close()
	dto := PigeonDto{
		Priority:   event.Priority(),
		Status:     event.Status,
		Endpoint:   event.Endpoint,
		Note:       event.Note(),
		Metric:     event.Metric(),
		Tags:       utils.SortedTags(event.PushedTags),
		Func:       event.Func(),
		LeftValue:  utils.ReadableFloat(event.LeftValue),
		Operator:   event.Operator(),
		RightValue: utils.ReadableFloat(event.RightValue()),
		EventTime:  event.FormattedTime(),
		IP:         ip,
		Domain:     hostgroup,
	}
	bs, err := json.Marshal(dto)
	if err != nil {
		log.Error("json marshal PigeonDto fail:", err)
		return
	}
	_, err = rc.Do("LPUSH", queue, string(bs))
	if err != nil {
		log.Error("LPUSH redis", queue, "fail:", err, "dto:", string(bs))
	}
}

func writePigeon(event *cmodel.Event, ip, hostgroup string) {
	if event == nil {
		return
	}
	pigeon := GeneratePigeon(event, ip, hostgroup)
	if pigeon != nil {
		redi.WritePigeonModel(pigeon)
	}

}

func getIPAndGroup(hostname string) (ip, group string) {
	group, err1 := g.GetHostGroup(hostname)
	ip, err2 := g.GetHostIP(hostname)
	if err1 != nil || group == "" || err2 != nil || ip == "" {
		host_info, err := g.GetHostInfoFromCMDB(hostname)
		if err != nil {
			if group == "" {
				group = "Unknown"
			}
		} else {
			if ip == "" {
				ip = host_info.IP
			}
			if group == "" {
				group = host_info.Domain
			}
		}
	}
	return
}

func consumeToPigeon(eventDetail *eventmodel.EventDetail) {
	if eventDetail == nil {
		return
	}
	event := eventDetail.Event
	ip := eventDetail.IP
	hostgroup := eventDetail.GrpName
	if hostgroup == "" {
		hostgroup = "Unknown"
	}

	if isNeedPigeonHighCombine(event.Priority()) {
		//send to redis, prepare for high combine
		queue := g.Config().Redis.PigeonHighQueue
		writePigeonDto(event, ip, hostgroup, queue)
	} else if isNeedPigeonLowCombine(event.Priority()) {
		//send to redis, prepare for low combine
		queue := g.Config().Redis.PigeonLowQueue
		writePigeonDto(event, ip, hostgroup, queue)
	} else {
		//don't need to combine
		writePigeon(event, ip, hostgroup)
	}
}
