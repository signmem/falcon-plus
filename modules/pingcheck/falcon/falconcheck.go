package falcon

import (
	"encoding/json"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/tools"
	"io/ioutil"
)

type FalconHostInfo struct {
	Hostname		string		`json:"hostname"`
	Ip 				string 		`json:"ip"`
}


func GetFalconHost(hostname string) (hostip string, err error) {
	// use to check hostname
	// GET  /api/vip/host/:hostname
	// example: http://falcon-api.vip.vip.com/api/vip/host/test-m3hqj.vclound.com
	// return  "ipaddr": string

	httpUrl := g.Config().Falcon.Url
	falconCheckHostApi := httpUrl + "/api/vip/host"
	falconCheckParams := "/" + hostname

	respBody, err := tools.HttpApiGet(falconCheckHostApi, falconCheckParams, "falcon")
	if err != nil {
		g.Logger.Errorf("GetFalconHost() access HttpApiGet() error with api: %v, " +
			"params %v ", falconCheckHostApi, falconCheckParams)
		return "", err
	}

	responseBodyHost, err := ioutil.ReadAll(respBody)
	defer respBody.Close()
	if err != nil {
		g.Logger.Errorf("[WARN] GetFalconHost() get hostname: %s error: %s ", hostname, err)
		return "", err
	}

	var hostInfo FalconHostInfo
	err = json.Unmarshal( responseBodyHost, &hostInfo )

	if err != nil {
		return "", err
	}


	return hostInfo.Ip, nil
}

func DeleteFlaonHost(hostname string) (err error){
	httpUrl := g.Config().Falcon.Url
	falconDeleteHostApi := httpUrl + "/api/vip/host"
	falconDeleteParams := "/" + hostname

	_, err = tools.HttpApiDelete(falconDeleteHostApi, falconDeleteParams, "falcon")
	if err != nil {
		g.Logger.Errorf("DeleteFlaonHost() access HttpApiDelete() error with api: %v, " +
			"params %v ", falconDeleteHostApi, falconDeleteParams)
		return  err
	}

	return nil

}

