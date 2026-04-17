package cmdb

import (
	"encoding/json"
	"github.com/signmem/falcon-plus/modules/status-collectd/g"
	"github.com/signmem/falcon-plus/modules/status-collectd/http"
	"io/ioutil"
)


func GetCMDBServerName(ipaddr string) (hostname string) {

	hostname = "none"
	api := "/server/query"
	query := "ip=" + ipaddr

	serverInfo, err := cmdbApiQuery(api, query)
	if err != nil {
		return
	}

	if  len(serverInfo.Object) > 0 {
		hostname = serverInfo.Object[0].ServerName
	}

	return hostname
}


func cmdbApiQuery(api string, query string ) (mqAppStruct  CmdbTotalObject, err error) {
	//  access cmdb api
	//  ???????????
	// params:
	//  api: "/app/query"   "/server/query"
	//  query: "name=" + hostgroup  or "app_name=" + hostgroup
	//
	// return CmdbTotalObject struct

	// httpUrl := g.Config().CmdbConfig.Server
	// sysName := g.Config().CmdbConfig.SysName
	// token := g.Config().CmdbConfig.Token

	httpUrl := g.Config().Cmdb.Url
	sysName := g.Config().Cmdb.SysName
	token := g.Config().Cmdb.Token

	params := "?sys_name=" + sysName + "&key=" + token + "&" + query

	cmdbHostGroupApi := httpUrl + api


	response, err :=  http.HttpApiGet(cmdbHostGroupApi, params, )
	if err != nil {
		g.Logger.Errorf("CmdbAppQuery() %s err ", query)
		return mqAppStruct, err
	}

	responseBody, err := ioutil.ReadAll(response)
	defer response.Close()

	if err != nil {
		g.Logger.Errorf("CmdbAppQuery() %s response err ", query)
		return mqAppStruct, err
	}

	err = json.Unmarshal(responseBody, &mqAppStruct)

	if err != nil {
		return mqAppStruct, err
	}

	return mqAppStruct, nil
}
