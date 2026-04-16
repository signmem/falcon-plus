package  http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)


func configSnmpRoutes() {

	http.HandleFunc("/upload/snmp/metrics",
		func(w http.ResponseWriter, r *http.Request){

			if r.Method != http.MethodPost {
				http.Error(w, "POST method support only.", http.StatusMethodNotAllowed)
				return
			}

			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				http.Error(w, "json format support only.", http.StatusUnsupportedMediaType)
				return
			}

			var snmpjson g.SnmpServer
			decoder := json.NewDecoder(r.Body)
			// decoder.DisallowUnknownFields()

			if err := decoder.Decode(&snmpjson); err != nil {
				log.Printf("[ERROR] configSnmpRoutes() json decode err: %s", err)
				http.Error(w, "json format not valid: "+err.Error(), http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			if snmpjson.SnmpFile == "" {
				http.Error(w, "snmpfile json value not valid", http.StatusBadRequest)
				return
			}

			if fileExists(snmpjson.SnmpFile) != true {
				http.Error(w, "snmpfile json not exists", http.StatusBadRequest)
				return
			}

			if snmpjson.SnmpInfo.Ipaddr == "" {
				http.Error(w, "snmpdetail.ipaddr value not valid", http.StatusBadRequest)
				return
			}

			if snmpjson.SnmpInfo.Port == 0 {
				snmpjson.SnmpInfo.Port = 161
			}

			if snmpjson.SnmpInfo.HostName == "" {
				snmpjson.SnmpInfo.HostName, _ = os.Hostname()
			}



			if snmpjson.SnmpInfo.Community == "" {
				snmpjson.SnmpInfo.Community = "public"
			}

			if snmpjson.SnmpInfo.Retry == 0 || snmpjson.SnmpInfo.Retry > 3 {
				snmpjson.SnmpInfo.Retry = 3
			}

			if snmpjson.SnmpInfo.Timeout == 0 || snmpjson.SnmpInfo.Retry > 30 {
				snmpjson.SnmpInfo.Retry = 30
			}

			if snmpjson.SnmpInfo.Step == 0 {
				snmpjson.SnmpInfo.Step = 60
			}

			snmpoid, err := loadSnmpJsonFile(snmpjson.SnmpFile)

			if err != nil {
				http.Error(w, "snmpfile json decode false", http.StatusBadRequest)
				return
			}

			runSnmapCheck(snmpjson, snmpoid)

			w.Write([]byte("snmp metrics send ok."))

		})
}


func runSnmapCheck(snmpServerDict g.SnmpServer, snmpoid g.SnmpOID) {

        var metricsValue float64
        err := snmapProgram(snmpServerDict, snmpoid)

        if err != nil {
                return
        }

        metricsValue = 1
        // 监控成功，则上报 falcon agent.alive = 1
        genSnmpMetricAlive(snmpServerDict, metricsValue)
        return
}


func snmapProgram(SnmpDict g.SnmpServer, snmpoid g.SnmpOID) (err error){

        // totalSnmpMetric 私有变量
        var totalSnmpMetric []*model.MetricValue

        debug := SnmpDict.Debug

        // range oids in cfg file (snmpget 用法)
        oidsmap := snmpoid.Oids

        for _, idsGroup := range oidsmap {

                // fix get metrics from global variable to private variable
                snmpGetMetrics, err := snmpGet(SnmpDict, "", idsGroup)

                if err != nil {
                        continue
                }

                totalSnmpMetric = append(totalSnmpMetric, snmpGetMetrics...)
        }

        // snmpwalk 用法

        oidwalk := snmpoid.OidWalks

        for _, oidWalkMap := range oidwalk {

                // 通过 walk 获取当前监控个数并创建新 tag
                //var oidWalkTag []OidStruct
                oidWalkTag, err := walkMakeTag(SnmpDict.SnmpInfo.Ipaddr, oidWalkMap.TagOid, debug)

                if err != nil {
                        log.Printf("[ERROR] spip check %s, oid %s\n",
                        	SnmpDict.SnmpInfo.Ipaddr, oidWalkMap.TagOid)
                        continue
                }

                // 读取 cfg check 配置 (后面用于遍历的 oid)
                oidWalkCheck := oidWalkMap.Check

                tagLeft :=  oidWalkMap.TagName

                if len(oidWalkTag) > 0 {
                        for _, getWalkTags := range oidWalkTag {
                                num := getWalkTags.WalkNum

                                // 创建 tag
                                tagRight := getWalkTags.WalkReturn
                                tagFull := tagLeft + tagRight

                                for _, walkCheck := range  oidWalkCheck {

                                        // 重组 oidmap 用于 snmp get
                                        var walkOidMap  g.OIDMAP
                                        // regroup walkOidMap
                                        walkCheckOid := walkCheck.OID + "." + num

                                        walkOidMap.OID =  walkCheckOid
                                        walkOidMap.Alias = walkCheck.Alias
                                        walkOidMap.Type = walkCheck.Type

                                        // 利用重组的 oid map 进行 snmpget 操作
                                        snmpMetrics, err := snmpGet(SnmpDict, tagFull, walkOidMap)

                                        if err != nil {
                                                continue
                                        }

                                        totalSnmpMetric = append(totalSnmpMetric, snmpMetrics...)
                                }

                        }
                }

                if len(totalSnmpMetric) == 0 {
                        errmsg := errors.New("server:%s, metric total len is 0")
                        return errmsg
                }

                if SnmpDict.Debug == true {
					log.Printf("[DEBUG] server: %s, metric total len: %d\n",
						SnmpDict.SnmpInfo.Ipaddr, len(totalSnmpMetric))

                        for _, metric := range totalSnmpMetric {
							log.Printf("[DEBUG]: %s\n", metric.String())
                        }
                }

                if len(totalSnmpMetric) > 0  {
					var resp model.TransferResponse
					g.SendMetrics(totalSnmpMetric, &resp)
                }

        }
        return nil
}


func snmpGet(SnmpServerDict g.SnmpServer, tagName string, idsGroup g.OIDMAP) (snmpMetrics []*model.MetricValue, err error) {

	var snmpVersion gosnmp.SnmpVersion

	version := SnmpServerDict.SnmpInfo.Version

	if version == 0 || version == 2 || version > 3 {
		snmpVersion = gosnmp.Version2c
	}

	if version == 1 {
		snmpVersion = gosnmp.Version1
	}

	if version == 3 {
		snmpVersion = gosnmp.Version3
	}


	gosnmp.Default.Target = SnmpServerDict.SnmpInfo.Ipaddr
	gosnmp.Default.Port = SnmpServerDict.SnmpInfo.Port
	gosnmp.Default.Version = snmpVersion
	gosnmp.Default.Community = SnmpServerDict.SnmpInfo.Community
	gosnmp.Default.Retries = SnmpServerDict.SnmpInfo.Retry
	gosnmp.Default.Timeout = time.Duration( SnmpServerDict.SnmpInfo.Timeout)  * time.Second
	err = gosnmp.Default.Connect()

	if err != nil {
		log.Printf("[ERROR] snmp get connect error:%s", err)
		return snmpMetrics, err
	}

	defer gosnmp.Default.Conn.Close()

	// must be initial metrics  !!!
	metrics := new(model.MetricValue)

	metrics.Endpoint = SnmpServerDict.SnmpInfo.HostName

	if len(tagName) != 0 {
		metrics.Tags = tagName
	}

	metrics.Timestamp = time.Now().Unix()
	metrics.Step = SnmpServerDict.SnmpInfo.Step

	var idsmap []string
	idsmap = append(idsmap, idsGroup.OID)

	result, err := gosnmp.Default.Get(idsmap)

	if err != nil {
		log.Printf("[ERROR] snmp Get() ids:%s, alias:%s, err:%s",
			idsGroup.OID, idsGroup.Alias, err)
		return snmpMetrics, err
	}

	for _, variables := range result.Variables {

		// g.Logger.Infof("oid: %s, nmae: %s", variables.Name, idsDetail.Name)

		metrics.Metric = idsGroup.Alias
		metrics.Type = idsGroup.Type

		switch variables.Type {

		case gosnmp.OctetString:
			// g.Logger.Infof("string: %v", string(variables.Value.([]byte)))
			valueString := string(variables.Value.([]byte))
			metrics.Value, _ = strconv.ParseFloat(valueString, 64)
		default:
			// g.Logger.Infof("num: %d", gosnmp.ToBigInt(variables.Value))
			value := gosnmp.ToBigInt(variables.Value).Int64()
			metrics.Value =  float64(value)
		}

		snmpMetrics = append(snmpMetrics, metrics)
	}
	time.Sleep(time.Millisecond * 200 )
	return snmpMetrics,nil
}

func walkMakeTag(address string, oids string, debug bool) (oidwalktag []g.OidStruct, err error) {

        gosnmp.Default.Target = address
        gosnmp.Default.Timeout = time.Duration(10 * time.Second)
        err = gosnmp.Default.Connect()

        if err != nil {
                log.Printf("[ERROR] snmp walk connect error:%s\n", err)
                return oidwalktag, err
        }

        defer gosnmp.Default.Conn.Close()

        // err = gosnmp.Default.BulkWalk(oids, walkValue)
        allPduReport, err := gosnmp.Default.BulkWalkAll(oids)

        if err != nil {
			log.Printf("[ERROR] snmp BulkWalkAll get error:%s\n", err)
                return oidwalktag, err
        }

        oidwalktag, err = walkValueToTag(allPduReport, debug)

        if err != nil {
			log.Printf("[ERROR] snmp %s walk %s error:%s\n", address, oids, err)
                return oidwalktag, err
        }

        return oidwalktag, nil
}

func walkValueToTag(report []gosnmp.SnmpPDU, debug bool) (totalOidTag []g.OidStruct, err error) {

        // g.Logger.Infof("walk name: %s", pdu.Name)
        for _, pdu := range report {

                var oidTag g.OidStruct

                lastOID := pdu.Name[strings.LastIndex(pdu.Name, ".")+1:]
                oidTag.WalkNum = lastOID

                switch pdu.Type {
                case gosnmp.OctetString:
                        b := pdu.Value.([]byte)
                        oidTag.WalkReturn = string(b)
                }

                if len(oidTag.WalkNum) == 0 || len(oidTag.WalkReturn) == 0 {
                	if debug == true {
						msg := fmt.Sprintf("[ERROR] snmpWalk get Tag error %s\n",
							pdu.Name)
						log.Printf(msg)
						continue
					}
                }

                totalOidTag = append(totalOidTag, oidTag)
        }

        return totalOidTag,nil
}

func genSnmpMetricAlive(snmpServerDict g.SnmpServer, value float64) (err error) {

	// metrics must be initial

	metrics := &model.MetricValue{
		Type:      "GAUGE",
		Step:      snmpServerDict.SnmpInfo.Step,
		Timestamp: time.Now().Unix(),
		Endpoint:  snmpServerDict.SnmpInfo.HostName,

		Value:     value,
	}

	var sendMetrics []*model.MetricValue

	metricSnmp := *metrics
	metricSnmp.Metric = "snmpd.alive"
	sendMetrics = append(sendMetrics, &metricSnmp)
	metricAlive := *metrics
	metricAlive.Metric = "agent.alive"
	sendMetrics = append(sendMetrics, &metricAlive)

	var resp model.TransferResponse
	g.SendMetrics(sendMetrics, &resp)

	return nil
}




func loadSnmpJsonFile(filePath string) (snmpoid g.SnmpOID, err error) {


	if _, err := os.Stat(filePath); err == nil {

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("[ERROR] snmp jsonfile %s read error\n", filePath)
			return snmpoid, err
		}

		err = json.Unmarshal(content, &snmpoid)

		if err != nil {
			fmt.Printf("cal metrics jsonfile %s json format error\n", filePath)
			return snmpoid, err
		}


	}

	return snmpoid, nil
}

func LoadSnmpOIDFile (filePath string) (validMetric g.SnmpOID, err error) {

	if _, err := os.Stat(filePath); err == nil {

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("[ERROR] loadjsonfile %s read error\n", filePath)
			return validMetric, err
		}

		err = json.Unmarshal(content, &validMetric)

		if err != nil {
			log.Printf("[ERROR] loadjsonfile %s json format error\n", filePath)
			return validMetric, err
		}

	} else {
		log.Printf("[ERROR] loadjsonfile %s not exists\n", filePath)
		return validMetric, err
	}
	return validMetric, nil
}
