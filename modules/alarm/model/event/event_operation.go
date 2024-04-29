package event

import (
	"fmt"
	"time"

	"database/sql"

	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	coommonModel "github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/common/utils"
	"github.com/signmem/falcon-plus/modules/alarm/g"
)

const timeLayout = "2006-01-02 15:04:05"

type hostGroupInfo struct {
	Host_id   int64
	Host_ip   string
	Buss_name string
	Group_id  int
	Grp_name  string
}

type EventDetail struct {
	GrpName string
	IP      string
	Event   *coommonModel.Event
}

func insertEvent(q orm.Ormer, eve *coommonModel.Event) (res sql.Result, err error) {
	var status int
	if status = 0; eve.Status == "OK" {
		status = 1
	}
	sqltemplete := `INSERT INTO events (
		event_caseId,
		step,
		cond,
		status,
		timestamp
	) VALUES(?,?,?,?,?)`
	res, err = q.Raw(
		sqltemplete,
		eve.Id,
		eve.CurrentStep,
		fmt.Sprintf("%v %v %v", eve.LeftValue, eve.Operator(), eve.RightValue()),
		status,
		time.Unix(eve.EventTime, 0).Format(timeLayout),
	).Exec()

	if err != nil {
		log.Errorf("insert event to db fail, error:%v", err)
	} else {
		lastid, _ := res.LastInsertId()
		log.Debug("insert event to db succ, last_insert_id:", lastid)
	}
	return
}

func getHostGroupInfo(event *coommonModel.Event, isFirst bool) *hostGroupInfo {
	q := orm.NewOrm()
	var host_id int64
	var host_ip string
	var grp_id int
	var grp_name string
	var buss_name string

	bOnlyRedis := false
	info := &hostGroupInfo{}

	//owner为realtime的告警app tag为domain
	v, ok := event.PushedTags["owner"]
	if ok && v == "realtime" {
		
		v, ok = event.PushedTags["host_"]
		if ok {
			info.Grp_name = v
		}
		
		//plugin上报数据变更，realtime docker告警域名固定为gd12-vrc-k8s-node
		//info.Grp_name = "gd12-vrc-k8s-node"
		info.Buss_name = "大数据平台/实时计算平台"
		return info
	}

	err := q.Raw(`select host_id, ip, grp_id, grp_name from falcon_portal.host,
        falcon_portal.grp_host, falcon_portal.grp
        where host.id=grp_host.host_id and grp.id=grp_host.grp_id and hostname = ?`,
		event.Endpoint).QueryRow(&host_id, &host_ip, &grp_id, &grp_name)
	if err != nil {
		log.Println(err.Error())
		info.Host_id = getHostId(event.Endpoint)
	} else {
		info.Host_id = host_id
		info.Host_ip = host_ip
		info.Group_id = grp_id
		info.Grp_name = grp_name
		log.Debugf("getHostGroupInfo, host info: %d, %s, %d, %s, %s", info.Host_id, info.Host_ip, info.Group_id, info.Grp_name, info.Buss_name)
	}

	if !isFirst || event.Priority() >= 4 {
		//非首次更新,event_case表中有且step大于1，以及p4以上，只获取redis缓存中信息
		bOnlyRedis = true
	}

	if info.Host_ip == "" {
		//如果host表中找不到ip，到缓存或cmdb中查找ip
		hostInfo, err := getHostInfo(event.Endpoint, bOnlyRedis)
		if err == nil {
			info.Host_ip = hostInfo.IP
			if info.Grp_name == "" {
				//没有group信息，使用缓存或cmdb中获取的group name
				info.Grp_name = hostInfo.Domain
			}
			info.Buss_name = hostInfo.Buss_name
		} else {
			//缓存或cmdb中获取失败，event_cases表中获取ip，group name和buss_name
			err = q.Raw(`select ip, grp_name, buss_name from event_cases where id = ?`, event.Id).QueryRow(&host_ip, &grp_name, &buss_name)
			if err == nil {
				if host_ip != "" {
					info.Host_ip = host_ip
				}
				if grp_name != "" {
					info.Grp_name = grp_name
				}
				if buss_name != "" {
					info.Buss_name = buss_name
				}
				log.Debugf("getHostGroupInfo from event_cases: %s, %s, %s", info.Host_ip, info.Grp_name, info.Buss_name)
			}
		}
	} else if info.Grp_name == "" {
		//host表中获取到ip信息,但没有group信息，使用缓存或cmdb中获取的group name
		hostInfo, err := getHostInfo(event.Endpoint, bOnlyRedis)
		if err == nil {
			//没有group信息，使用缓存或cmdb中的group name
			info.Grp_name = hostInfo.Domain
			info.Buss_name = hostInfo.Buss_name
		} else {
			//缓存或cmdb中获取失败，event_cases表中获取group name和buss_name
			err = q.Raw(`select grp_name, buss_name from event_cases where id = ?`, event.Id).QueryRow(&grp_name, &buss_name)
			if err == nil {
				if grp_name != "" {
					info.Grp_name = grp_name
				}
				if buss_name != "" {
					info.Buss_name = buss_name
				}
				log.Debugf("getHostGroupInfo from event_cases: %s, %s", info.Grp_name, info.Buss_name)
			}
		}
	} else {
		if isFirst {
			hostInfo, err := getHostInfo(event.Endpoint, bOnlyRedis)
			if err == nil {
				info.Buss_name = hostInfo.Buss_name
			} else {
				err = q.Raw(`select buss_name from event_cases where id = ?`, event.Id).QueryRow(&buss_name)
				if err == nil {
					if buss_name != "" {
						info.Buss_name = buss_name
					}
					log.Debugf("getHostGroupInfo from event_cases: %s", info.Buss_name)
				}
			}
		}
	}

	return info
}

func getHostInfo(endpoint string, bOnlyRedis bool) (*g.HostObject, error) {
	rc := g.RedisConnPoolCDMBCache.Get()
	defer func() {
		if rc != nil {
			rc.Close()
		}
	}()
	hostInfo := &g.HostObject{}
	log.Debugf("getHostInfo: %s, %t", endpoint, bOnlyRedis)
	//redis缓存中获取host信息
	reply, err := redis.String(rc.Do("GET", "server:"+endpoint))

	if err != nil {
		log.Errorf("get host info from redis failed: %v", err)
	} else {
		var serverObj map[string]interface{}

		if err = json.Unmarshal([]byte(reply), &serverObj); err == nil {
			value, ok := serverObj["server_name"].(string)
			if ok {
				hostInfo.Hostname = value
			}
			value, ok = serverObj["ip"].(string)
			if ok {
				hostInfo.IP = value
			}
			value, ok = serverObj["domain_name"].(string)
			if ok {
				hostInfo.Domain = value
			}
			value, ok = serverObj["buss_name"].(string)
			if ok {
				hostInfo.Buss_name = value
			}
			log.Debugf("getHostInfo from redis, host info: %s, %s, %s, %s", hostInfo.Hostname, hostInfo.IP, hostInfo.Domain, hostInfo.Buss_name)
			//redis中获取到host信息，支持返回
			return hostInfo, nil
		} else {
			log.Errorf("get host info from redis fail: %v", err)
		}
	}

	//缓存中获取失败，从cmdb中获取信息
	if !bOnlyRedis {
		hostInfo, err = g.GetHostInfoFromCMDB(endpoint)
		if err == nil {
			log.Debugf("getHostInfo from cmdb, host info: %s, %s, %s, %s", hostInfo.Hostname, hostInfo.IP, hostInfo.Domain, hostInfo.Buss_name)
		}
		return hostInfo, err
	} else {
		// 仅取缓存
		return &g.HostObject{}, err
	}
}

func getHostId(endpoint string) int64 {
	q := orm.NewOrm()
	var host_id int64

	err := q.Raw(`select id from falcon_portal.host where hostname = ?`, endpoint).QueryRow(&host_id)
	if err != nil {
		log.Errorf("get host [%s] id error: %v", endpoint, err)
		return 0
	}
	return host_id
}

func InsertEvent(eve *coommonModel.Event, changeIgnore, sendMoreMax bool) (*EventDetail, bool) {
	q := orm.NewOrm()
	var event []EventCases
	q.Raw("select * from event_cases where id = ?", eve.Id).QueryRows(&event)
	var sqlLog sql.Result
	var errRes error
	//modified by vincent.zhang for ignore process
	needConsume := true
	info := &hostGroupInfo{}
	eventDetail := &EventDetail{}
	eventDetail.Event = eve

	log.Debugf("events: %v", eve)
	log.Debugf("expression is null: %v", eve.Expression == nil)
	if len(event) == 0 {
		//新的告警需要进行ip，group，buss_name插入
		info = getHostGroupInfo(eve, true)
		//create cases
		sqltemplete := `INSERT INTO event_cases (
					id,
					host_id,
					ip,
					endpoint,
					buss_name,
					grp_id,
					grp_name,
					metric,
					func,
					cond,
					note,
					max_step,
					current_step,
					priority,
					status,
					timestamp,
					update_at,
					tpl_creator,
					expression_id,
					strategy_id,
					template_id
					) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

		tpl_creator := ""
		if eve.Tpl() != nil {
			tpl_creator = eve.Tpl().Creator
		}
		sqlLog, errRes = q.Raw(
			sqltemplete,
			eve.Id,
			info.Host_id,
			info.Host_ip,
			eve.Endpoint,
			info.Buss_name,
			info.Group_id,
			info.Grp_name,
			counterGen(eve.Metric(), utils.SortedTags(eve.PushedTags)),
			eve.Func(),
			//cond
			fmt.Sprintf("%v %v %v", eve.LeftValue, eve.Operator(), eve.RightValue()),
			eve.Note(),
			eve.MaxStep(),
			eve.CurrentStep,
			eve.Priority(),
			eve.Status,
			//start_at
			time.Unix(eve.EventTime, 0).Format(timeLayout),
			//update_at
			time.Unix(eve.EventTime, 0).Format(timeLayout),
			tpl_creator,
			eve.ExpressionId(),
			eve.StrategyId(),
			//template_id
			eve.TplId()).Exec()

	} else {
		sqltemplete := `UPDATE event_cases SET
				update_at = ?,
				max_step = ?,
				current_step = ?,
				note = ?,
				cond = ?,
				status = ?,
				func = ?,
				priority = ?,
				tpl_creator = ?,
				expression_id = ?,
				strategy_id = ?,
				template_id = ?,
				grp_id = ?,
				grp_name = ?`
		//reopen case
		//modified by vincent.zhang for ignore process
		/*
			if event[0].ProcessStatus == "resolved" || event[0].ProcessStatus == "ignored" {
				sqltemplete = fmt.Sprintf("%v ,process_status = '%s', process_note = %d", sqltemplete, "unresolved", 0)
			}
		*/
		if sendMoreMax == false {
			//事件和记录同为problem，并且记录已达到最大发送次数，不发送
			if eve.Status == "PROBLEM" && event[0].Status == "PROBLEM" {
				if event[0].CurrentStep >= event[0].MaxStep && eve.MaxStep() <= event[0].MaxStep {
					needConsume = false
					log.Debugf("Max step event don't consume, event: %s\n", eve.String())
					return eventDetail, needConsume
				}
			}
		}
		if event[0].ProcessStatus == "resolved" {
			sqltemplete = fmt.Sprintf("%v ,process_status = '%s', process_note = %d", sqltemplete, "unresolved", 0)
		} else if event[0].ProcessStatus == "ignored" {
			if changeIgnore == true {
				sqltemplete = fmt.Sprintf("%v ,process_status = '%s', process_note = %d", sqltemplete, "unresolved", 0)
			} else {
				//ok事件改变process_status
				if eve.Status == "OK" || event[0].Status == "OK" {
					sqltemplete = fmt.Sprintf("%v ,process_status = '%s', process_note = %d", sqltemplete, "unresolved", 0)
				} else {
					needConsume = false
					log.Debugf("Ignored event don't consume, event: %s\n", eve.String())
					//return needConsume
				}
			}
		}
		//modified end

		tpl_creator := ""
		if eve.Tpl() != nil {
			tpl_creator = eve.Tpl().Creator
		}
		if eve.CurrentStep == 1 {
			//告警的第一次需要进行ip，group，buss_name更新
			info = getHostGroupInfo(eve, true)
			sqltemplete = fmt.Sprintf("%v ,ip = '%s', buss_name = '%s'", sqltemplete, info.Host_ip, info.Buss_name)

			//update start time of cases
			sqltemplete = fmt.Sprintf("%v ,timestamp = ? WHERE id = ?", sqltemplete)
			sqlLog, errRes = q.Raw(
				sqltemplete,
				time.Unix(eve.EventTime, 0).Format(timeLayout),
				eve.MaxStep(),
				eve.CurrentStep,
				eve.Note(),
				fmt.Sprintf("%v %v %v", eve.LeftValue, eve.Operator(), eve.RightValue()),
				eve.Status,
				eve.Func(),
				eve.Priority(),
				tpl_creator,
				eve.ExpressionId(),
				eve.StrategyId(),
				eve.TplId(),
				info.Group_id,
				info.Grp_name,
				time.Unix(eve.EventTime, 0).Format(timeLayout),
				eve.Id,
			).Exec()
		} else {
			//非第一次告警需要进行group更新
			info = getHostGroupInfo(eve, false)

			sqltemplete = fmt.Sprintf("%v WHERE id = ?", sqltemplete)
			sqlLog, errRes = q.Raw(
				sqltemplete,
				time.Unix(eve.EventTime, 0).Format(timeLayout),
				eve.MaxStep(),
				eve.CurrentStep,
				eve.Note(),
				fmt.Sprintf("%v %v %v", eve.LeftValue, eve.Operator(), eve.RightValue()),
				eve.Status,
				eve.Func(),
				eve.Priority(),
				tpl_creator,
				eve.ExpressionId(),
				eve.StrategyId(),
				eve.TplId(),
				info.Group_id,
				info.Grp_name,
				eve.Id,
			).Exec()
		}
	}
	log.Debug(fmt.Sprintf("%v, %v", sqlLog, errRes))
	//insert case
	//modified by vincent.zhang for don't insert event of ignore and p4 level
	if needConsume && eve.Priority() < 4 {
		insertEvent(q, eve)
	}

	eventDetail.GrpName = info.Grp_name
	eventDetail.IP = info.Host_ip
	//added by vincent.zhang for ignore process
	return eventDetail, needConsume
}

func counterGen(metric string, tags string) (mycounter string) {
	mycounter = metric
	if tags != "" {
		mycounter = fmt.Sprintf("%s/%s", metric, tags)
	}
	return
}

func DeleteEventOlder(before time.Time, limit int) {
	t := before.Format(timeLayout)
	sqlTpl := `delete from events where timestamp<? limit ?`
	q := orm.NewOrm()
	resp, err := q.Raw(sqlTpl, t, limit).Exec()
	if err != nil {
		log.Errorf("delete event older than %v fail, error:%v", t, err)
	} else {
		affected, _ := resp.RowsAffected()
		log.Debugf("delete event older than %v, rows affected:%v", t, affected)
	}
}
