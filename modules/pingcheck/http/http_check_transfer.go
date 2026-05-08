package http

import (
	"fmt"
	httpclient "github.com/signmem/falcon-plus/common/http"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"strings"
	"time"
)

func CheckTransfer() {
	transferServers := g.Config().Transfer

	for {
		httpCount := 0

		for _, server := range transferServers.Servers {
			healthUrl := "http://" + server + "/health"
			resp, err := httpclient.HttpApiGet(healthUrl, "", "")
			if err != nil {
				fmt.Println(err)
				continue
			}
			body := string(resp)

			if strings.Contains(body, "ok") {
				httpCount += 1
			}
		}
		if len(transferServers.Servers) == 0 {
			g.TransferCheck = false
		} else 	if httpCount > len(transferServers.Servers) / 2 {
			g.TransferCheck = true
		} else {
			g.TransferCheck = false
		}
		time.Sleep(time.Duration(transferServers.Interval) * time.Second)
	}
}



