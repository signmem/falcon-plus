package cron

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	cmodel "github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/alarm/g"
	eventmodel "github.com/signmem/falcon-plus/modules/alarm/model/event"
)

func ReadHighEvent() {
	queues := g.Config().Redis.HighQueues
	changeIgnore := g.Config().ChangeIgnore
	sendMoreMax := g.Config().SendMoreMax
	if len(queues) == 0 {
		return
	}

	for {
		//modified by vincent.zhang for ignore process
		event, bConsume, err := popEvent(queues, changeIgnore, sendMoreMax)
		if err != nil {
			time.Sleep(time.Second)
			continue
		} else if bConsume == false {
			continue
		}
		consume(event, true)
	}
}

func ReadLowEvent() {
	queues := g.Config().Redis.LowQueues
	changeIgnore := g.Config().ChangeIgnore
	sendMoreMax := g.Config().SendMoreMax
	if len(queues) == 0 {
		return
	}

	for {
		//modified by vincent.zhang for ignore process
		event, bConsume, err := popEvent(queues, changeIgnore, sendMoreMax)
		if err != nil {
			time.Sleep(time.Second)
			continue
		} else if bConsume == false {
			continue
		}
		consume(event, false)
	}
}

func popEvent(queues []string, changeIgnore, sendMoreMax bool) (*eventmodel.EventDetail, bool, error) {

	count := len(queues)

	params := make([]interface{}, count+1)
	for i := 0; i < count; i++ {
		params[i] = queues[i]
	}
	// set timeout 0
	//params[count] = 0
	//modify by vincent.zhang for BRPOP can not return when redis error
	params[count] = 60

	rc := g.RedisConnPool.Get()
	defer rc.Close()
	//added by vincent.zhang for ignore process
	needConsume := true

	reply, err := redis.Strings(rc.Do("BRPOP", params...))
	if err != nil {
		//log.Errorf("get alarm event from redis fail: %v", err)
		//modify by vincent.zhang for BRPOP can not return when redis error
		if err != redis.ErrNil {
			log.Errorf("get alarm event from redis fail: %v", err)
		}
		//end
		//modified by vincent.zhang for ignore process
		//return nil, err
		return nil, needConsume, err
	}

	var event cmodel.Event
	err = json.Unmarshal([]byte(reply[1]), &event)
	if err != nil {
		log.Errorf("parse alarm event fail: %v", err)
		//modified by vincent.zhang for ignore process
		//return nil, err
		return nil, needConsume, err
	}

	log.Debugf("pop event: %s", event.String())

	//insert event into database
	//eventmodel.InsertEvent(&event, changeIgnore)
	//modified by vincent.zhang for ignore process
	var eventDetail *eventmodel.EventDetail
	eventDetail, needConsume = eventmodel.InsertEvent(&event, changeIgnore, sendMoreMax)
	// events no longer saved in memory

	return eventDetail, needConsume, nil
}
