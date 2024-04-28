package cron

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	//add by vincent.zhang fro pigeon
	log "github.com/Sirupsen/logrus"
	"github.com/toolkits/net/httplib"
	"regexp"
	"sort"
	//"strconv"
	"time"
	//"strings"

	//"github.com/open-falcon/falcon-plus/modules/alarm/api"
	amodel "github.com/open-falcon/falcon-plus/modules/alarm/model"
	//add end
)

const (
	PIGEON_KEY              = "99a35ace8952f662cad672d2f6da5754"
	PIGEON_SOURE_NAME       = "vip-falcon"
	PIGEON_DEFAULT_FID      = 9560
	PIGEON_DEFAULT_AVC      = "000-000"
	PIGEON_FID_CREATE_ADMIN = "vincent.zhang"
)

func BuildCommonSMSContent(event *model.Event) string {
	return fmt.Sprintf(
		"[L%d][%s][%s][%s %s %s %s%s%s][%s]",
		event.Priority(),
		event.Status,
		event.Endpoint,
		event.Note(),
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.FormattedTime(),
	)
}

func BuildCommonIMContent(event *model.Event) string {
	return fmt.Sprintf(
		"[L%d][%s][%s][][%s %s %s %s %s%s%s][O%d %s]",
		event.Priority(),
		event.Status,
		event.Endpoint,
		event.Note(),
		event.Func(),
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.CurrentStep,
		event.FormattedTime(),
	)
}

func BuildCommonMailContent(event *model.Event) string {
	link := g.Link(event)
	return fmt.Sprintf(
		"%s\r\nP%d\r\nEndpoint:%s\r\nMetric:%s\r\nTags:%s\r\n%s: %s%s%s\r\nNote:%s\r\nMax:%d, Current:%d\r\nTimestamp:%s\r\n%s\r\n",
		event.Status,
		event.Priority(),
		event.Endpoint,
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		event.Func(),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.Note(),
		event.MaxStep(),
		event.CurrentStep,
		event.FormattedTime(),
		link,
	)
}

func GenerateSmsContent(event *model.Event) string {
	return BuildCommonSMSContent(event)
}

func GenerateMailContent(event *model.Event) string {
	return BuildCommonMailContent(event)
}

func GenerateIMContent(event *model.Event) string {
	return BuildCommonIMContent(event)
}

//add pigeon by vincent.zhang
func buildPigeonSubject(event *model.Event, ip, hostgroup string) string {
	if event == nil {
		return ""
	}
	//add by vincent.zhang for dealing with endpoint without ip
	host := ""
	if ip == "" {
		host = event.Endpoint
	} else {
		host = ip
	}

	return fmt.Sprintf(
		"[%s] %s %s",
		hostgroup,
		host,
		event.Note(),
	)
}

func buildPigeonExtagrLegend(event *model.Event, ip, hostgroup string) string {
	if event == nil {
		return ""
	}
	//add by vincent.zhang for dealing with endpoint without ip
	host := ""
	if ip == "" {
		host = event.Endpoint
	} else {
		host = ip
	}

	return fmt.Sprintf(
		"[%s] %s",
		hostgroup,
		host,
	)
}

func buildPigeonExtagrTitil(event *model.Event) string {
	if event == nil {
		return ""
	}

	titleInfo := fmt.Sprintf(
		"%s",
		event.Note(),
	)
	return strings.Replace(titleInfo,"%","",-1)
}

func buildPigeonM3Body(event *model.Event) string {

	var newM3Body amodel.M3Body
	tags := utils.SortedTags(event.PushedTags)
	if tags == "" {
		counter = event.Metric()
	} else {
		counter = event.Metric() + "/" + tags
	}

	counter_type, _ := g.GetMetricType(event.Endpoint, counter)

	metricInfo       := buildPigeonExtagrMetric(event)

	if counter_type == "DERIVE" {
		newM3Body.Metric      = fmt.Sprintf("%s%s%s","increase(", metricInfo, "[5m])")
	}else{
		newM3Body.Metric      = fmt.Sprintf("%s",metricInfo)
	}

	newM3Body.DatasourceID = 3
	newM3Body.From         = time.Now().Unix() - 3600
	newM3Body.Step         = 60
	newM3Body.Source       = "81a7b791e01f4b899dc2ceaa047b25d7"
	newM3Body.To           = time.Now().Unix()
	newM3Body.Type         = "query_range"
	newM3BodyJson, _ := json.Marshal(newM3Body)
	return fmt.Sprintf(string(newM3BodyJson))
}

func buildPigeonM3Chart(event *model.Event, ip, hostgroup string) amodel.M3Chart {
	var newM3Chart amodel.M3Chart
	newM3Chart.Legend = buildPigeonExtagrLegend(event, ip, hostgroup)
	newM3Chart.Title =  buildPigeonExtagrTitil(event)
	return newM3Chart
}

func buildPigeonExtagrMetric(event *model.Event) string {
	if event == nil {
		return ""
	}

	m3dbMetricName :=  fmt.Sprintf(
		"%s",
		strings.Replace(event.Metric(),".","_",-1),
	)

	var defaultTags string
	var oneLineTags []string
	var m3dbSearchMetric string
	for key, val := range event.PushedTags {
		defaultTags = fmt.Sprintf("%s=\"%s\"", key, val)
		oneLineTags = append ( oneLineTags, defaultTags)
	}

	hostinfo := "host_name"
	hostvalue := event.Endpoint
	hostinfo = fmt.Sprintf("%s=\"%s\"",hostinfo, hostvalue)

	oneLineTags = append ( oneLineTags, hostinfo)

	formatTags := fmt.Sprintf(strings.Join(oneLineTags,","))
	m3dbSearchMetric = fmt.Sprintf("%s{%s}", m3dbMetricName, formatTags)
	return m3dbSearchMetric
}

func buildPigeonM3Extargs (event *model.Event, ip, hostgroup string) *amodel.ExtArg {

	var newM3Value amodel.M3Value
	var newM3ValueSlice []amodel.M3Value
	newM3Value.Body = buildPigeonM3Body(event)
	newM3Value.Chart = buildPigeonM3Chart(event,ip,hostgroup)
	newM3Value.URL = "http://m3.api.vip.com/api/getdata"
	newM3ValueSlice = append(newM3ValueSlice, newM3Value )
	m3ValueJson, _ := json.Marshal(newM3ValueSlice)
	m3Values := fmt.Sprintf(string(m3ValueJson))

	m3ExtraArgs := amodel.ExtArg{
		Name : "wx_chart_url",
		Value: m3Values,
	}

	return &m3ExtraArgs

}

func buildPigeonSMS(event *model.Event, ip, hostgroup string) string {
	if event == nil {
		return ""
	}

	tags := utils.SortedTags(event.PushedTags)

	//modified by vincent.zhang for host without ip
	host := ""
	if ip == "" {
		host = fmt.Sprintf("[%s] %s", hostgroup, event.Endpoint)
	} else {
		host = fmt.Sprintf("[%s] %s(%s)", hostgroup, event.Endpoint, ip)
	}

	if tags == "" {
		return fmt.Sprintf(
			"%s %s, 当前值：%s",
			host,
			event.Note(),
			utils.ReadableFloat(event.LeftValue),
		)
	} else {
		return fmt.Sprintf(
			"%s %s %s, 当前值：%s",
			host,
			event.Note(),
			tags,
			utils.ReadableFloat(event.LeftValue),
		)
	}
}

func buildPigeonMessage(event *model.Event, ip, hostgroup string) string {
	if event == nil {
		return ""
	}
	tags := utils.SortedTags(event.PushedTags)

	//modified by vincent.zhang for host without ip
	host := ""
	if ip == "" {
		host = fmt.Sprintf("[%s] %s", hostgroup, event.Endpoint)
	} else {
		host = fmt.Sprintf("[%s] %s(%s)", hostgroup, event.Endpoint, ip)
	}

	if tags == "" {
		return fmt.Sprintf(
			"%s %s %s %s%s%s, 当前值：%s",
			host,
			event.Note(),
			event.Metric(),
			event.Func(),
			event.Operator(),
			utils.ReadableFloat(event.RightValue()),
			utils.ReadableFloat(event.LeftValue),
		)
	} else {
		return fmt.Sprintf(
			"%s %s %s/%s %s%s%s, 当前值：%s",
			host,
			event.Note(),
			event.Metric(),
			tags,
			event.Func(),
			event.Operator(),
			utils.ReadableFloat(event.RightValue()),
			utils.ReadableFloat(event.LeftValue),
		)
	}
}

func removePoolTag(tags string) (newTag string) {

	// use to remove tags include (pool=xxxx)

	var nameTag []string
	nameSplit := strings.Split(tags, ",")
	for _, spName := range nameSplit {
		if strings.Contains(spName, "pool=") {
			continue
		} else {
			nameTag = append(nameTag, spName)
		}
	}
	newTag = strings.Join(nameTag[:], ",")
	return
}


//modified by vincent.zhang for falcon view chart
func buildPigeonTransfer(hostname, metric, tags string) string {
	if tags == "" {
		counter = metric
	} else {
		newTags := removePoolTag(tags)
		counter = metric + "/" + newTags
	}
	counter_type, _ := g.GetMetricType(hostname, counter)
	if counter_type != "" {
		counter = counter + "_type_" + counter_type
	}
	return fmt.Sprintf("http://falcon-view.vip.vip.com/charts?endpoints=%s&metrics=%s", hostname, counter)
}

func buildPigeonFidName(metric, tags, funcName, operator, rightValue string) string {
	func_name := funcName
	if tags == "" {
		return fmt.Sprintf(
			"%s[%s%s%s]",
			metric,
			func_name,
			operator,
			rightValue,
		)
	} else {
		return fmt.Sprintf(
			"%s/%s[%s%s%s]",
			metric,
			tags,
			func_name,
			operator,
			rightValue,
		)
	}
}

func createPigeonFid(name, creator string, note string, priority int) int64 {
	if name == "" {
		log.Warningf("pigeon fid name is empty\n")
		return PIGEON_DEFAULT_FID
	}
	url := g.Config().Pigeon.FidAddr
	if url == "" {
		log.Warningf("pigeon fid addr is empty\n")
		return PIGEON_DEFAULT_FID
	}

	create_admin := creator
	if create_admin == "" {
		create_admin = PIGEON_FID_CREATE_ADMIN
	}
	//added by vincent.zhang for deciding fid type
	fid_type := "class_type_40" //class_type_10 网络 ,class_type_20 数据库 ,class_type_30 业务 ,class_type_40 系统
	//class_type_50 硬件,class_type_60 应用,class_type_70 用户 ,class_type_80 动环
	//老版本 1:业务 2:网络 3:硬件 4:数据库 5:系统 6:应用 7:用户
	ok, reg_err := regexp.MatchString("^hardware", name)
	if reg_err != nil {
		log.Errorf("regexp hardware error: name:%s, error:%s", name, reg_err.Error())
	} else {
		if ok {
			fid_type = "class_type_50" //硬件
		}
	}

	req := httplib.Get(url).SetTimeout(3*time.Second, 20*time.Second)
	req.Param("name", name)
	req.Param("creator", create_admin)
	req.Param("sys_token", PIGEON_KEY)
	req.Param("source_name", PIGEON_SOURE_NAME)
	req.Param("is_deal", "0")
	req.Param("class_type", fid_type)
	req.Param("rul_name", note)
	req.Param("note", "created by falcon-alarm")
	if priority < 3 {
		req.Param("notice_type", "2")
	} else {
		req.Param("notice_type", "1")
	}

	resp := PigeonFidResponse{}
	err := req.ToJson(&resp)

	log.Debugf("get pigeon fid name:%s, resp:%v", name, resp)
	if err != nil {
		log.Errorf("get Pigeon Fid fail, name:%s, error:%s", name, err.Error())
		return PIGEON_DEFAULT_FID
	}
	//code==9为fid已存在
	if resp.Success == false && resp.Code != 9 {
		log.Errorf("get pigeon fid fail, name:%s, message:%s", name, resp.Message)
	} else {
		return resp.Object.Fid
	}
	return PIGEON_DEFAULT_FID
}

func getPigeonFidAndVAC(event *model.Event) (int64, string) {
	if event == nil {
		log.Warningf("get pigeon fid parameter is nil\n")
		return PIGEON_DEFAULT_FID, PIGEON_DEFAULT_AVC
	}

	var fid int64
	var err error
	var avc string
	url := g.Config().Pigeon.FidAddr
	/*
		if url == "" {
			log.Warningf("pigeon fid addr is empty\n")
			return PIGEON_DEFAULT_FID, PIGEON_DEFAULT_AVC
		}
	*/
	strategy_id := event.StrategyId()
	if strategy_id > 0 {
		fid, avc, err = g.GetFidAndAVCFromDB(strategy_id, "strategy")
		if err != nil {
			if url != "" {
				strategy := event.Strategy
				fid_name := buildPigeonFidName(strategy.Metric, utils.SortedTags(strategy.Tags), strategy.Func, strategy.Operator, utils.ReadableFloat(strategy.RightValue))
				fid = createPigeonFid(fid_name, strategy.Tpl.Creator, strategy.Note, strategy.Priority)
			} else {
				log.Warningf("pigeon fid addr is empty\n")
				fid = PIGEON_DEFAULT_FID
			}
			if fid != PIGEON_DEFAULT_FID {
				g.UpdateFidToDB(strategy_id, fid, "strategy")
			}
		}
	} else {
		expression_id := event.ExpressionId()
		if expression_id > 0 {
			fid, avc, err = g.GetFidAndAVCFromDB(expression_id, "expression")
			if err != nil {
				if url != "" {
					expression := event.Expression
					fid_name := buildPigeonFidName(expression.Metric, utils.SortedTags(expression.Tags), expression.Func, expression.Operator, utils.ReadableFloat(expression.RightValue))
					fid = createPigeonFid(fid_name, "", expression.Note, expression.Priority)
				} else {
					log.Warningf("pigeon fid addr is empty\n")
					fid = PIGEON_DEFAULT_FID
				}
				if fid != PIGEON_DEFAULT_FID {
					g.UpdateFidToDB(expression_id, fid, "expression")
				}
			}
		} else {
			fid = PIGEON_DEFAULT_FID
		}
	}
	if avc == "" {
		avc = PIGEON_DEFAULT_AVC
	}
	return fid, avc
}

//added by vincent.zhang for new pigon api with ext_tags
func buildPigeonExtTags(tags map[string]string) []*amodel.ExtArg {
	if tags == nil {
		return []*amodel.ExtArg{}
	}

	size := len(tags)

	if size == 0 {
		return []*amodel.ExtArg{}
	}
	ret := make([]*amodel.ExtArg, size)
	if size == 1 {
		for k, v := range tags {
			extArg := amodel.ExtArg{
				Name:  k,
				Value: v,
			}
			ret[0] = &extArg
			return ret
		}
	}

	keys := make([]string, size)
	i := 0
	for k := range tags {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	for j, key := range keys {
		extArg := amodel.ExtArg{
			Name:  key,
			Value: tags[key],
		}
		ret[j] = &extArg
	}
	return ret
}

func buildPigeonExtArgs(tags map[string]string, right_value string) []*amodel.ExtArg {
	ret := buildPigeonExtTags(tags)
	rightvalueArg := amodel.ExtArg{
		Name:  "right_value",
		Value: right_value,
	}
	ret = append(ret, &rightvalueArg)
	return ret
}

func buildPigeonM3Transfer(hostname string, metric string) string {

	m3metric := strings.Replace(metric, ".", "_", -1)
	return fmt.Sprintf("https://m3.vip.vip.com/v3/dashboard/panel/view/255/0?refresh=1m&orgId=1&from=now-3h&to=now&var-host_name=" + hostname +  "&var-metric=" + m3metric)

}


func GeneratePigeon(event *model.Event, ip, hostgroup string) *amodel.Pigeon {
	if event == nil {
		return nil
	}
	// tags := utils.SortedTags(event.PushedTags)
	// rightValue := utils.ReadableFloat(event.RightValue())
	// transfer := buildPigeonTransfer(event.Endpoint, event.Metric(), tags)
	// fidname := BuildPigeonFidName(event.Metric(), tags, event.Func(), event.Operator(), rightValue)

	transfer := buildPigeonM3Transfer(event.Endpoint, event.Metric())
	fid, avc := getPigeonFidAndVAC(event)
	m3ExtArgs := buildPigeonM3Extargs(event, ip, hostgroup)
	defaultArgs := buildPigeonExtArgs(event.PushedTags, utils.ReadableFloat(event.RightValue()))

	defaultArgs = append(defaultArgs, m3ExtArgs)

	pigeon := amodel.Pigeon{
		Fid:       fid,
		AlarmCode: avc,
		Status:    event.Status,
		Value:     utils.ReadableFloat(event.LeftValue),
		Subject:   buildPigeonSubject(event, ip, hostgroup),
		Sms:       buildPigeonSMS(event, ip, hostgroup),
		Message:   buildPigeonMessage(event, ip, hostgroup),
		Priority:  event.Priority(),
		Host:      ip,
		HostName:  event.Endpoint,
		Domain:    hostgroup,
		Transfer:  transfer,
		AlarmTime: event.FormattedTime(),
		// ExtArgs:   buildPigeonExtArgs(event.PushedTags, utils.ReadableFloat(event.RightValue())),
		ExtArgs:   defaultArgs,
	}
	
	return &pigeon
}
