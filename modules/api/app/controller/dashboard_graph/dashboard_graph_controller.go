package dashboard_graph

import (
	"fmt"
	"github.com/signmem/falcon-plus/modules/api/config"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cutils "github.com/signmem/falcon-plus/common/utils"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	m "github.com/signmem/falcon-plus/modules/api/app/model/dashboard"
)

type APITmpGraphCreateReqData struct {
	Endpoints []string `json:"endpoints" binding:"required"`
	Counters  []string `json:"counters" binding:"required"`
}

func DashboardTmpGraphCreate(c *gin.Context) {
	var inputs APITmpGraphCreateReqData
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	es := inputs.Endpoints
	cs := inputs.Counters
	sort.Strings(es)
	sort.Strings(cs)

	es_string := strings.Join(es, TMP_GRAPH_FILED_DELIMITER)
	cs_string := strings.Join(cs, TMP_GRAPH_FILED_DELIMITER)
	ck := cutils.Md5(es_string + ":" + cs_string)
	//modified by vincent.zhang for id increasing quickly
	tmp_graph := m.DashboardTmpGraph{}
	dt := db.Dashboard.Table("tmp_graph").Where("ck=?", ck).First(&tmp_graph)
	if dt.Error != nil || tmp_graph.ID == 0 {
		dt = db.Dashboard.Exec("insert ignore into `tmp_graph` (endpoints, counters, ck) values(?, ?, ?) on duplicate key update time_=?", es_string, cs_string, ck, time.Now())
		if dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			return
		}

		dt = db.Dashboard.Table("tmp_graph").Where("ck=?", ck).First(&tmp_graph)
		if dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			return
		}
	}
	/*
		dt := db.Dashboard.Exec("insert ignore into `tmp_graph` (endpoints, counters, ck) values(?, ?, ?) on duplicate key update time_=?", es_string, cs_string, ck, time.Now())
		if dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			return
		}

		tmp_graph := m.DashboardTmpGraph{}
		dt = db.Dashboard.Table("tmp_graph").Where("ck=?", ck).First(&tmp_graph)
		if dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			return
		}
	*/
	h.JSONR(c, map[string]int{"id": int(tmp_graph.ID)})
}

func DashboardTmpGraphQuery(c *gin.Context) {
	id := c.Param("id")

	tmp_graph := m.DashboardTmpGraph{}
	dt := db.Dashboard.Table("tmp_graph").Where("id = ?", id).First(&tmp_graph)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	es := strings.Split(tmp_graph.Endpoints, TMP_GRAPH_FILED_DELIMITER)
	cs := strings.Split(tmp_graph.Counters, TMP_GRAPH_FILED_DELIMITER)

	ret := map[string][]string{
		"endpoints": es,
		"counters":  cs,
	}

	h.JSONR(c, ret)
}

type APIGraphCreateReqData struct {
	ScreenId   int      `json:"screen_id" binding:"required"`
	Title      string   `json:"title" binding:"required"`
	Endpoints  []string `json:"endpoints" binding:"required"`
	Counters   []string `json:"counters" binding:"required"`
	TimeSpan   int      `json:"timespan"`
	GraphType  string   `json:"graph_type"`
	Method     string   `json:"method"`
	Position   int      `json:"position"`
	FalconTags string   `json:"falcon_tags"`
}

func DashboardGraphCreate(c *gin.Context) {
	var inputs APIGraphCreateReqData
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	es := inputs.Endpoints
	cs := inputs.Counters
	sort.Strings(es)
	sort.Strings(cs)
	es_string := strings.Join(es, TMP_GRAPH_FILED_DELIMITER)
	cs_string := strings.Join(cs, TMP_GRAPH_FILED_DELIMITER)

	d := m.DashboardGraph{
		Title:     inputs.Title,
		Hosts:     es_string,
		Counters:  cs_string,
		ScreenId:  int64(inputs.ScreenId),
		TimeSpan:  inputs.TimeSpan,
		GraphType: inputs.GraphType,
		Method:    inputs.Method,
		Position:  inputs.Position,
	}
	if d.TimeSpan == 0 {
		d.TimeSpan = 3600
	}
	if d.GraphType == "" {
		d.GraphType = "h"
	}

	graph := m.DashboardGraph{}

	/*qt := db.Dashboard.Table("dashboard_graph").Where("screen_id = ? AND title = ? AND counters = ?",
	inputs.ScreenId, inputs.Title, cs_string).Find(&graph)*/

	qt := db.Dashboard.Table("dashboard_graph").Where("screen_id = ? AND title = ?", inputs.ScreenId, inputs.Title).Find(&graph)
	if qt.RowsAffected != 0 {
		config.Logger.Infof("%T, %v", qt.RowsAffected, qt.RowsAffected)
		h.JSONR(c, badstatus, fmt.Sprintf("Title %v is already exist in the screen %v!", inputs.Title, inputs.ScreenId))
		return
	}

	dt := db.Dashboard.Table("dashboard_graph").Create(&d)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	var lid []int
	dt = db.Dashboard.Table("dashboard_graph").Raw("select LAST_INSERT_ID() as id").Pluck("id", &lid)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	aid := lid[0]

	h.JSONR(c, map[string]int{"id": aid})

}

type APIGraphUpdateReqData struct {
	ScreenId   int      `json:"screen_id"`
	Title      string   `json:"title"`
	Endpoints  []string `json:"endpoints"`
	Counters   []string `json:"counters"`
	TimeSpan   int      `json:"timespan"`
	GraphType  string   `json:"graph_type"`
	Method     string   `json:"method"`
	Position   int      `json:"position"`
	FalconTags string   `json:"falcon_tags"`
}

func DashboardGraphUpdate(c *gin.Context) {
	id := c.Param("id")
	gid, err := strconv.Atoi(id)
	if err != nil {
		h.JSONR(c, badstatus, "invalid graph id")
		return
	}

	var inputs APIGraphUpdateReqData
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	d := m.DashboardGraph{}

	if len(inputs.Endpoints) != 0 {
		es := inputs.Endpoints
		sort.Strings(es)
		es_string := strings.Join(es, TMP_GRAPH_FILED_DELIMITER)
		d.Hosts = es_string
	}
	if len(inputs.Counters) != 0 {
		cs := inputs.Counters
		sort.Strings(cs)
		cs_string := strings.Join(cs, TMP_GRAPH_FILED_DELIMITER)
		d.Counters = cs_string
	}
	if inputs.Title != "" {
		d.Title = inputs.Title
	}
	if inputs.ScreenId != 0 {
		d.ScreenId = int64(inputs.ScreenId)
	}
	if inputs.TimeSpan != 0 {
		d.TimeSpan = inputs.TimeSpan
	}
	if inputs.GraphType != "" {
		d.GraphType = inputs.GraphType
	}
	if inputs.Method != "" {
		d.Method = inputs.Method
	}
	if inputs.Position != 0 {
		d.Position = inputs.Position
	}
	if inputs.FalconTags != "" {
		d.FalconTags = inputs.FalconTags
	}

	graph := m.DashboardGraph{}
	dt := db.Dashboard.Table("dashboard_graph").Model(&graph).Where("id = ?", gid).Updates(d)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, map[string]int{"id": gid})

}

func DashboardGraphGet(c *gin.Context) {
	id := c.Param("id")
	gid, err := strconv.Atoi(id)
	if err != nil {
		h.JSONR(c, badstatus, "invalid graph id")
		return
	}

	graph := m.DashboardGraph{}
	dt := db.Dashboard.Table("dashboard_graph").Where("id = ?", gid).First(&graph)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	es := strings.Split(graph.Hosts, TMP_GRAPH_FILED_DELIMITER)
	cs := strings.Split(graph.Counters, TMP_GRAPH_FILED_DELIMITER)

	h.JSONR(c, map[string]interface{}{
		"graph_id":    graph.ID,
		"title":       graph.Title,
		"endpoints":   es,
		"counters":    cs,
		"screen_id":   graph.ScreenId,
		"graph_type":  graph.GraphType,
		"timespan":    graph.TimeSpan,
		"method":      graph.Method,
		"position":    graph.Position,
		"falcon_tags": graph.FalconTags,
	})

}

func DashboardGraphDelete(c *gin.Context) {
	id := c.Param("id")
	gid, err := strconv.Atoi(id)
	if err != nil {
		h.JSONR(c, badstatus, "invalid graph id")
		return
	}

	graph := m.DashboardGraph{}
	dt := db.Dashboard.Table("dashboard_graph").Where("id = ?", gid).Delete(&graph)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, map[string]int{"id": gid})

}

func DashboardGraphGetsByScreenID(c *gin.Context) {
	id := c.Param("screen_id")
	sid, err := strconv.Atoi(id)
	if err != nil {
		h.JSONR(c, badstatus, "invalid screen id")
		return
	}
	limit := c.DefaultQuery("limit", "500")

	graphs := []m.DashboardGraph{}
	dt := db.Dashboard.Table("dashboard_graph").Where("screen_id = ?", sid).Limit(limit).Find(&graphs)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	ret := []map[string]interface{}{}
	for _, graph := range graphs {
		es := strings.Split(graph.Hosts, TMP_GRAPH_FILED_DELIMITER)
		cs := strings.Split(graph.Counters, TMP_GRAPH_FILED_DELIMITER)

		r := map[string]interface{}{
			"graph_id":    graph.ID,
			"title":       graph.Title,
			"endpoints":   es,
			"counters":    cs,
			"screen_id":   graph.ScreenId,
			"graph_type":  graph.GraphType,
			"timespan":    graph.TimeSpan,
			"method":      graph.Method,
			"position":    graph.Position,
			"falcon_tags": graph.FalconTags,
		}
		ret = append(ret, r)
	}

	h.JSONR(c, ret)
}

type APIDashboardGraphGetsInputs struct {
	ScreenID string `json:"screen_id" form:"screen_id"`
	Title    string `json:"title" form:"title"`
}

func DashboardGraphGetsByScreenAndTitle(c *gin.Context) {
	inputs := APIDashboardGraphGetsInputs{}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	if inputs.ScreenID == "" && inputs.Title == "" {
		h.JSONR(c, http.StatusBadRequest, "screen_id and title are all missing")
		return
	}

	sid, err := strconv.Atoi(inputs.ScreenID)
	if err != nil {
		h.JSONR(c, badstatus, "invalid screen id")
		return
	}

	graphs := []m.DashboardGraph{}
	dt := db.Dashboard.Table("dashboard_graph").Where("screen_id = ? and title = ?", sid, inputs.Title).Find(&graphs)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	ret := []map[string]interface{}{}
	for _, graph := range graphs {
		es := strings.Split(graph.Hosts, TMP_GRAPH_FILED_DELIMITER)
		cs := strings.Split(graph.Counters, TMP_GRAPH_FILED_DELIMITER)

		r := map[string]interface{}{
			"graph_id":    graph.ID,
			"title":       graph.Title,
			"endpoints":   es,
			"counters":    cs,
			"screen_id":   graph.ScreenId,
			"graph_type":  graph.GraphType,
			"timespan":    graph.TimeSpan,
			"method":      graph.Method,
			"position":    graph.Position,
			"falcon_tags": graph.FalconTags,
		}
		ret = append(ret, r)
	}

	h.JSONR(c, ret)
}
