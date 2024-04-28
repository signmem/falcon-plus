package cmdb

import (

	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/tools"
	"encoding/json"
	"io/ioutil"
)



func CmdbApiQuery(api string, query string ) (mqAppStruct  CmdbTotalObject, err error) {
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
	response, err :=  tools.HttpApiGet(cmdbHostGroupApi, params, "")
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



