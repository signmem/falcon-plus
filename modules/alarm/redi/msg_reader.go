package redi

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/signmem/falcon-plus/modules/alarm/g"
	"github.com/signmem/falcon-plus/modules/alarm/model"
)

const (
	IM_QUEUE_NAME     = "/im"
	SMS_QUEUE_NAME    = "/sms"
	MAIL_QUEUE_NAME   = "/mail"
	PIGEON_QUEUE_NAME = "/pigeon" //add by vincent.zhang for pigeon
)

func PopAllSms() []*model.Sms {
	ret := []*model.Sms{}
	queue := SMS_QUEUE_NAME

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				log.Error(err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var sms model.Sms
		err = json.Unmarshal([]byte(reply), &sms)
		if err != nil {
			log.Error(err, reply)
			continue
		}

		ret = append(ret, &sms)
	}

	return ret
}

func PopAllIM() []*model.IM {
	ret := []*model.IM{}
	queue := IM_QUEUE_NAME

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				log.Error(err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var im model.IM
		err = json.Unmarshal([]byte(reply), &im)
		if err != nil {
			log.Error(err, reply)
			continue
		}

		ret = append(ret, &im)
	}

	return ret
}

func PopAllMail() []*model.Mail {
	ret := []*model.Mail{}
	queue := MAIL_QUEUE_NAME

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				log.Error(err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var mail model.Mail
		err = json.Unmarshal([]byte(reply), &mail)
		if err != nil {
			log.Error(err, reply)
			continue
		}

		ret = append(ret, &mail)
	}

	return ret
}

//add by vincent.zhang for pigeon
func PopAllPigeon() []*model.Pigeon {
	ret := []*model.Pigeon{}
	queue := PIGEON_QUEUE_NAME

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				log.Error(err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var pigeon model.Pigeon
		err = json.Unmarshal([]byte(reply), &pigeon)
		if err != nil {
			log.Error(err, reply)
			continue
		}

		ret = append(ret, &pigeon)
	}

	return ret
}

//end
