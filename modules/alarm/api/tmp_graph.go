package api

import (
	"encoding/json"
	"fmt"
	"github.com/toolkits/net/httplib"
	"log"
	"time"

	"github.com/open-falcon/falcon-plus/modules/alarm/g"
)

type TmpGraphResp struct {
	Id int `json:"id"`
}

type APITmpGraphInput struct {
	Endpoints []string `json:"endpoints" binding:"required"`
	Counters  []string `json:"counters" binding:"required"`
}

func CreateTmpGraph(input *APITmpGraphInput) *TmpGraphResp {
	uri := fmt.Sprintf("%s/api/v1/dashboard/tmpgraph", g.Config().Api.PlusApi)
	req := httplib.Post(uri).SetTimeout(3*time.Second, 10*time.Second)
	token, _ := json.Marshal(map[string]string{
		"name": "falcon-alarm",
		"sig":  "default-token-used-in-server-side",
	})
	req.Header("Apitoken", string(token))
	req.Header("Content-Type", "application/json")

	b, err := json.Marshal(input)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	req.Body(b)

	var resp TmpGraphResp
	err = req.ToJson(&resp)

	if err != nil {
		log.Printf("http %s fail: %v", uri, err)
		return nil
	}
	return &resp
}
