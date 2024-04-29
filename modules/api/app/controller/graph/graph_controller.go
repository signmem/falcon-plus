package graph

import (
	"fmt"
	"github.com/signmem/falcon-plus/modules/api/config"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	cmodel "github.com/signmem/falcon-plus/common/model"
	cutils "github.com/signmem/falcon-plus/common/utils"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	f "github.com/signmem/falcon-plus/modules/api/app/model/falcon_portal"
	m "github.com/signmem/falcon-plus/modules/api/app/model/graph"
	"github.com/signmem/falcon-plus/modules/api/app/utils"
	grh "github.com/signmem/falcon-plus/modules/api/graph"
)

//added by vincent.zhang for screen bug of dashboard
type APIEndpointObjGetInputs struct {
	Endpoints []string `json:"endpoints" form:"endpoints"`
	Deadline  int64    `json:"deadline" form:"deadline"`
}

func EndpointObjGet(c *gin.Context) {
	inputs := APIEndpointObjGetInputs{
		Deadline: 0,
	}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	if len(inputs.Endpoints) == 0 {
		h.JSONR(c, http.StatusBadRequest, "endpoints missing")
		return
	}

	var result []m.Endpoint = []m.Endpoint{}
	dt := db.Graph.Table("endpoint").
		Where("endpoint in (?) and ts >= ?", inputs.Endpoints, inputs.Deadline).
		Scan(&result)
	if dt.Error != nil {
		h.JSONR(c, http.StatusBadRequest, dt.Error)
		return
	}

	endpoints := []map[string]interface{}{}
	for _, r := range result {
		endpoints = append(endpoints, map[string]interface{}{"id": r.ID, "endpoint": r.Endpoint, "ts": r.Ts})
	}

	h.JSONR(c, endpoints)
}

//added end

type APIEndpointRegexpQueryInputs struct {
	Q     string `json:"q" form:"q"`
	Label string `json:"tags" form:"tags"`
	Limit int    `json:"limit" form:"limit"`
}

func EndpointRegexpQuery(c *gin.Context) {
	inputs := APIEndpointRegexpQueryInputs{
		//set default is 500
		Limit: 500,
	}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	if inputs.Q == "" && inputs.Label == "" {
		h.JSONR(c, http.StatusBadRequest, "q and labels are all missing")
		return
	}

	labels := []string{}
	if inputs.Label != "" {
		labels = strings.Split(inputs.Label, ",")
	}
	qs := []string{}
	if inputs.Q != "" {
		qs = strings.Split(inputs.Q, " ")
	}

	var endpoint []m.Endpoint
	var endpoint_id []int
	var dt *gorm.DB
	if len(labels) != 0 {
		dt = db.Graph.Table("endpoint_counter").Select("distinct endpoint_id")
		for _, trem := range labels {
			dt = dt.Where(" counter like ? ", "%"+strings.TrimSpace(trem)+"%")
		}
		dt = dt.Limit(500).Pluck("distinct endpoint_id", &endpoint_id)
		if dt.Error != nil {
			h.JSONR(c, http.StatusBadRequest, dt.Error)
			return
		}
	}
	if len(qs) != 0 {
		dt = db.Graph.Table("endpoint").
			Select("endpoint, id")
		if len(endpoint_id) != 0 {
			dt = dt.Where("id in (?)", endpoint_id)
		}

		for _, trem := range qs {
			dt = dt.Where(" endpoint regexp ? ", strings.TrimSpace(trem))
		}
		dt.Limit(inputs.Limit).Scan(&endpoint)
	} else if len(endpoint_id) != 0 {
		dt = db.Graph.Table("endpoint").
			Select("endpoint, id").
			Where("id in (?)", endpoint_id).
			Limit(inputs.Limit).
			Scan(&endpoint)
	}
	if dt.Error != nil {
		h.JSONR(c, http.StatusBadRequest, dt.Error)
		return
	}

	endpoints := []map[string]interface{}{}
	for _, e := range endpoint {
		endpoints = append(endpoints, map[string]interface{}{"id": e.ID, "endpoint": e.Endpoint})
	}

	h.JSONR(c, endpoints)
}

func EndpointCounterDistinct(c *gin.Context) {
	plan := c.DefaultQuery("plan", "full")

	currentTime := time.Now()
	var checkDate string
	if plan == "" || plan == "full" {
		oldTime := currentTime.AddDate(0, 0, -5)
		checkDate = oldTime.Format("2006-01-02")   // get data from 2days to now
	} else if plan == "append" {
		m, _ := time.ParseDuration("-120m")
		oldtime := currentTime.Add(m)
		checkDate = oldtime.Format("2006-01-02 15:04:05")
	}

	config.Logger.Debugf("EndpointCounterDistinct() check date from: %s", checkDate)
	var dt *gorm.DB
	var counters []string
	var counterReturn []string

	dt = db.Graph.Table("endpoint_counter").
		Select("distinct counter").
		Where("t_create > (?)", checkDate)
	dt = dt.Pluck("distinct counter", &counters)

	if dt.Error != nil {
		config.Logger.Errorf("db check error %s", dt.Error)
		h.JSONR(c, http.StatusBadRequest, dt.Error)
		return
	}

	if len(counters) == 0 {
		config.Logger.Debug("EndpointCounterDistinct() get zone metric.")
		h.JSONR(c,counterReturn)
		return
	}

	for _, counter := range counters {
		fitstString := strings.Split(counter, "/")

		if cutils.StringInSlice(fitstString[0], counterReturn) == false {
			counterReturn = append(counterReturn, fitstString[0])
		}
	}
	config.Logger.Debugf("EndpointCounterDistinct() counter length %d", len(counterReturn) )
	h.JSONR(c, counterReturn)
	return
}

func EndpointCounterRegexpQuery(c *gin.Context) {
	eid := c.DefaultQuery("eid", "")
	metricQuery := c.DefaultQuery("metricQuery", ".+")
	limitTmp := c.DefaultQuery("limit", "500")
	limit, err := strconv.Atoi(limitTmp)
	if err != nil {
		h.JSONR(c, http.StatusBadRequest, err)
		return
	}
	if eid == "" {
		h.JSONR(c, http.StatusBadRequest, "eid is missing")
	} else {
		eids := utils.ConverIntStringToList(eid)
		if eids == "" {
			h.JSONR(c, http.StatusBadRequest, "input error, please check your input info.")
			return
		} else {
			eids = fmt.Sprintf("(%s)", eids)
		}

		var counters []m.EndpointCounter
		dt := db.Graph.Table("endpoint_counter").Select("counter, step, type").Where(fmt.Sprintf("endpoint_id IN %s", eids))
		if metricQuery != "" {
			qs := strings.Split(metricQuery, " ")
			if len(qs) > 0 {
				for _, term := range qs {
					dt = dt.Where("counter regexp ?", strings.TrimSpace(term))
				}
			}
		}
		dt = dt.Limit(limit).Scan(&counters)
		if dt.Error != nil {
			h.JSONR(c, http.StatusBadRequest, dt.Error)
			return
		}

		countersResp := []interface{}{}
		for _, c := range counters {
			countersResp = append(countersResp, map[string]interface{}{
				"counter": c.Counter,
				"step":    c.Step,
				"type":    c.Type,
			})
		}
		h.JSONR(c, countersResp)
	}
	return
}

func EndpointCounterlistByIP(c *gin.Context) {
	ip := c.Params.ByName("ip")
	if ip == "" {
		h.JSONR(c, badstatus, "IP is missing")
		return
	}
	host := f.Host{}
	if dt := db.Falcon.Where("ip = ?", ip).Find(&host); dt.Error != nil {
		h.JSONR(c, badstatus, fmt.Sprintf("Host with IP %v does not found!", ip))
		return
	}

	endpoint := m.Endpoint{}
	if dt := db.Graph.Where("endpoint = ?", host.Hostname).Find(&endpoint); dt.Error != nil {
		h.JSONR(c, badstatus, fmt.Sprintf("Endpoint with name %v does not found!", host.Hostname))
		return
	}

	eid := endpoint.ID
	metricQuery := ".+"
	limit := 1000

	var counters []m.EndpointCounter
	dt := db.Graph.Table("endpoint_counter").Select("counter").Where("endpoint_id = ?", eid)
	if metricQuery != "" {
		qs := strings.Split(metricQuery, " ")
		if len(qs) > 0 {
			for _, term := range qs {
				dt = dt.Where("counter regexp ?", strings.TrimSpace(term))
			}
		}
	}
	dt = dt.Limit(limit).Scan(&counters)
	if dt.Error != nil {
		h.JSONR(c, http.StatusBadRequest, dt.Error)
		return
	}

	countersResp := []string{}
	for _, c := range counters {
		countersResp = append(countersResp, c.Counter)
	}

	h.JSONR(c, countersResp)
}

type APIQueryGraphDrawData struct {
	HostNames []string `json:"hostnames" binding:"required"`
	Counters  []string `json:"counters" binding:"required"`
	ConsolFun string   `json:"consol_fun" binding:"required"`
	StartTime int64    `json:"start_time" binding:"required"`
	EndTime   int64    `json:"end_time" binding:"required"`
	Step      int      `json:"step"`
}

func QueryGraphDrawData(c *gin.Context) {
	var inputs APIQueryGraphDrawData
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	respData := []*cmodel.GraphQueryResponse{}
	for _, host := range inputs.HostNames {
		for _, counter := range inputs.Counters {
			// TODO:cache step
			var step []int
			dt := db.Graph.Raw("select a.step from endpoint_counter as a, endpoint as b where b.endpoint = ? and a.endpoint_id = b.id and a.counter = ? limit 1", host, counter).Scan(&step)
			if dt.Error != nil || len(step) == 0 {
				continue
			}
			data, _ := fetchData(host, counter, inputs.ConsolFun, inputs.StartTime, inputs.EndTime, step[0])
			respData = append(respData, data)
		}
	}
	h.JSONR(c, respData)
}

type APIQueryLastPointInputs struct {
	Endpoints []string `json:"endpoints" binding:"required"`
	Counters  []string `json:"counters" binding:"required"`
}

func QueryGraphLastPoint(c *gin.Context) {
	var inputs APIQueryLastPointInputs
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	respData := []*cmodel.GraphLastResp{}

	for _, endpoint := range inputs.Endpoints {
		for _, counter := range inputs.Counters {
			param := cmodel.GraphLastParam{endpoint, counter}
			one_resp, err := grh.Last(param)
			if err != nil {
				config.Logger.Warningf("[WARN] query last point from graph fail: %s", err)
			} else {
				respData = append(respData, one_resp)
			}
		}
	}

	h.JSONR(c, respData)
}

func DeleteGraphEndpoint(c *gin.Context) {
	var inputs []string = []string{}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	type DBRows struct {
		Endpoint  string
		CounterId int
		Counter   string
		Type      string
		Step      int
	}
	rows := []DBRows{}
	dt := db.Graph.Raw(
		`select a.endpoint, b.id AS counter_id, b.counter, b.type, b.step from endpoint as a, endpoint_counter as b
		where b.endpoint_id = a.id
		AND a.endpoint in (?)`, inputs).Scan(&rows)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	var affected_counter int64 = 0
	var affected_endpoint int64 = 0

	if len(rows) > 0 {
		var params []*cmodel.GraphDeleteParam = []*cmodel.GraphDeleteParam{}
		for _, row := range rows {
			param := &cmodel.GraphDeleteParam{
				Endpoint: row.Endpoint,
				DsType:   row.Type,
				Step:     row.Step,
			}
			fields := strings.SplitN(row.Counter, "/", 2)
			if len(fields) == 1 {
				param.Metric = fields[0]
			} else if len(fields) == 2 {
				param.Metric = fields[0]
				param.Tags = fields[1]
			} else {
				config.Logger.Infof("invalid counter %s", row.Counter)
				continue
			}
			params = append(params, param)
		}
		grh.Delete(params)
	}

	tx := db.Graph.Begin()

	if len(rows) > 0 {
		var cids []int = make([]int, len(rows))
		for i, row := range rows {
			cids[i] = row.CounterId
		}

		dt = tx.Table("endpoint_counter").Where("id in (?)", cids).Delete(&m.EndpointCounter{})
		if dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			tx.Rollback()
			return
		}
		affected_counter = dt.RowsAffected

		dt = tx.Raw(`delete from tag_endpoint where endpoint_id in 
			(select id from endpoint where endpoint in (?))`, inputs)
		if dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			tx.Rollback()
			return
		}
	}

	dt = tx.Table("endpoint").Where("endpoint in (?)", inputs).Delete(&m.Endpoint{})
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	affected_endpoint = dt.RowsAffected
	tx.Commit()

	h.JSONR(c, map[string]int64{
		"affected_endpoint": affected_endpoint,
		"affected_counter":  affected_counter,
	})
}

type APIGraphDeleteCounterInputs struct {
	Endpoints []string `json:"endpoints" binding:"required"`
	Counters  []string `json:"counters" binding:"required"`
}

func DeleteGraphCounter(c *gin.Context) {
	var inputs APIGraphDeleteCounterInputs = APIGraphDeleteCounterInputs{}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	type DBRows struct {
		Endpoint  string
		CounterId int
		Counter   string
		Type      string
		Step      int
	}
	rows := []DBRows{}
	dt := db.Graph.Raw(`select a.endpoint, b.id AS counter_id, b.counter, b.type, b.step from endpoint as a, endpoint_counter as b
		where b.endpoint_id = a.id 
		AND a.endpoint in (?)
		AND b.counter in (?)`, inputs.Endpoints, inputs.Counters).Scan(&rows)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	if len(rows) == 0 {
		h.JSONR(c, map[string]int64{
			"affected_counter": 0,
		})
		return
	}

	var params []*cmodel.GraphDeleteParam = []*cmodel.GraphDeleteParam{}
	for _, row := range rows {
		param := &cmodel.GraphDeleteParam{
			Endpoint: row.Endpoint,
			DsType:   row.Type,
			Step:     row.Step,
		}
		fields := strings.SplitN(row.Counter, "/", 2)
		if len(fields) == 1 {
			param.Metric = fields[0]
		} else if len(fields) == 2 {
			param.Metric = fields[0]
			param.Tags = fields[1]
		} else {
			config.Logger.Errorf("invalid counter %s", row.Counter)
			continue
		}
		params = append(params, param)
	}
	grh.Delete(params)

	tx := db.Graph.Begin()
	var cids []int = make([]int, len(rows))
	for i, row := range rows {
		cids[i] = row.CounterId
	}

	dt = tx.Table("endpoint_counter").Where("id in (?)", cids).Delete(&m.EndpointCounter{})
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	affected_counter := dt.RowsAffected
	tx.Commit()

	h.JSONR(c, map[string]int64{
		"affected_counter": affected_counter,
	})
}

func fetchData(hostname string, counter string, consolFun string, startTime int64, endTime int64, step int) (resp *cmodel.GraphQueryResponse, err error) {
	qparm := grh.GenQParam(hostname, counter, consolFun, startTime, endTime, step)
	// log.Debugf("qparm: %v", qparm)
	resp, err = grh.QueryOne(qparm)
	if err != nil {
		config.Logger.Errorf("query graph got error: %s", err.Error())
	}
	return
}
