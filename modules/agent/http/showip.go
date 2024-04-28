package http

import (
	"net"
	"net/http"
)

func configShowIp() {

	http.HandleFunc("/v1/ipaddr", func(w http.ResponseWriter, r *http.Request) {

		var totalIP []string
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			RenderMsgJson(w, err.Error())
			return
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok  {
				if ipnet.IP.To4() != nil {
					ip := ipnet.IP.To4().String()
					totalIP = append( totalIP, ip)
				}
			}
		}


		RenderDataJson(w, map[string]interface{}{
			"ipaddr" : totalIP,
		})
	})




}
