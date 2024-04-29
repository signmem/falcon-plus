package g

import (
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus" //modified by vincent.zhang for pigeon
	"github.com/astaxie/beego/orm"
	"github.com/toolkits/net/httplib"
)

type HostObject struct {
	Hostname  string `json:"server_name"`
	IP        string `json:"ip"`
	Domain    string `json:"domain_name"`
	Buss_name string `json:"buss_name`
}

type CMDHostInfoResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Hosts   []*HostObject `json:"object"`
	Success bool          `json:"success"`
}

//modified by vincent.zhang for pigeon
func GetHostIP(hostname string) (string, error) {
	q := orm.NewOrm()
	var host_ip string

	err := q.Raw(`select ip from falcon_portal.host where hostname = ?`, hostname).QueryRow(&host_ip)
	if err != nil {
		log.Errorf("get host [%s] ip error: %v", hostname, err)
		return "", err
	}
	return host_ip, nil
}

func GetHostGroup(hostname string) (string, error) {
	q := orm.NewOrm()
	var host_group string

	err := q.Raw(`select a.grp_name from falcon_portal.grp a,falcon_portal.grp_host b, falcon_portal.host c where a.id=b.grp_id and b.host_id=c.id and c.hostname = ?`, hostname).QueryRow(&host_group)
	if err != nil {
		log.Errorf("get host [%s] hostgroup error: %v", hostname, err)
		return "", err
	}
	return host_group, nil
}

func GetHostInfoFromCMDB(hostname string) (*HostObject, error) {
	addr := Config().Api.CMDB
	if addr == "" {
		log.Errorf("GetHostInfoFromCMDB:cmdb api address is empty")
		return nil, errors.New("cmdb api address is empty.")
	}
	//url := fmt.Sprintf("%sserver/query?sys_name=%s&key=%s", addr, CMDB_APP, CMDB_KEY)
	url := addr + "server/query"
	req := httplib.Get(url).SetTimeout(5*time.Second, 30*time.Second)
	req.Param("sys_name", CMDB_APP)
	req.Param("key", CMDB_KEY)
	if net.ParseIP(hostname) == nil {
		//hostname is hostname
		req.Param("server_name", hostname)
	} else {
		//parameter is ip
		req.Param("ip", hostname)
	}
	resp := CMDHostInfoResponse{}
	err := req.ToJson(&resp)

	log.Debugf("GetHostInfoFromCMDB resp:%+v, hostname:%s, url:%s", resp, hostname, url)
	if err != nil {
		log.Errorf("GetHostInfoFromCMDB fail, error:%s", err.Error())
		return nil, err
	}
	if resp.Success == false {
		log.Errorf("GetHostInfoFromCMDB fail, message:%s", resp.Message)
		return nil, errors.New("cmdb api return false.")
	}
	if resp.Hosts == nil || len(resp.Hosts) <= 0 {
		log.Errorf("GetHostInfoFromCMDB return empty.")
		return nil, errors.New("cmdb api return empty.")
	}
	return resp.Hosts[0], nil
}

func GetFidAndAVCFromDB(id int, stype string) (int64, string, error) {
	q := orm.NewOrm()
	var fid int64
	avc := ""
	if stype == "strategy" {
		err := q.Raw(`select fid, pigeon_code from falcon_portal.strategy where id = ?`, id).QueryRow(&fid, &avc)
		if err != nil {
			log.Debugf("get strategy [%d] fid avc error: %v", id, err)
			return 0, "", err
		}
		if fid > 0 {
			return fid, avc, nil
		} else {
			return 0, avc, errors.New(fmt.Sprintf("strategy fid didn't stored,strategy id: %d", id))
		}
	} else if stype == "expression" {
		err := q.Raw(`select fid, pigeon_code from falcon_portal.expression where id = ?`, id).QueryRow(&fid, &avc)
		if err != nil {
			log.Debugf("get expression [%d] fid avc error: %v", id, err)
			return 0, "", err
		}
		if fid > 0 {
			return fid, avc, nil
		} else {
			return 0, avc, errors.New(fmt.Sprintf("expression fid didn't stored,strategy id: %d", id))
		}
	} else {
		err := errors.New("Get fid avc type is not support:" + stype)
		log.Errorf(err.Error())
		return 0, "", err
	}
}

func UpdateFidToDB(id int, fid int64, stype string) error {
	q := orm.NewOrm()
	if stype == "strategy" {
		sqlLog, err := q.Raw(`update falcon_portal.strategy set fid = ? where id = ?`, fid, id).Exec()
		if err != nil {
			log.Errorf("update strategy fid error: %v, %v", sqlLog, err)
			return err
		}
		return nil
	} else if stype == "expression" {
		sqlLog, err := q.Raw(`update falcon_portal.expression set fid = ? where id = ?`, fid, id).Exec()
		if err != nil {
			log.Errorf("update expression fid error: %v, %v", sqlLog, err)
			return err
		}
		return nil
	} else {
		err := errors.New("update fid type is not support:" + stype)
		log.Errorf(err.Error())
		return err
	}
}

func GetMetricType(hostname, counter string) (string, error) {
	q := orm.NewOrm()
	counter_type := ""

	err := q.Raw(`select type from graph.endpoint, graph.endpoint_counter where endpoint.id=endpoint_counter.endpoint_id and 
		endpoint = ? and counter = ?`, hostname, counter).QueryRow(&counter_type)
	if err != nil {
		log.Errorf("get [%s|%s] counter type error: %v", hostname, counter, err)
		return "", err
	}
	return counter_type, nil
}

/*
func GetHostIP(hostname string) *struct {
	Hostname string
	IP       string
} {
	q := orm.NewOrm()
	var host_name, host_ip string

	err := q.Raw(`select hostname, ip from falcon_portal.host where hostname = ?`,hostname).QueryRow(&host_name, &host_ip)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return &struct {
		Hostname string
		IP       string
	}{
		Hostname: host_name,
		IP:       host_ip,
	}
}


func GetHostGroup(hostname string) *struct {
        Hostgroup       string
} {
        q := orm.NewOrm()
        var host_group string

        err := q.Raw(`select a.grp_name from falcon_portal.grp a,falcon_portal.grp_host b, falcon_portal.host c where a.id=b.grp_id and b.host_id=c.id and c.hostname = ?`,hostname).QueryRow(&host_group)
        if err != nil {
                log.Println(err.Error())
                return nil
        }

        return &struct {
                Hostgroup string
        }{
                Hostgroup: host_group,
        }
}
*/
