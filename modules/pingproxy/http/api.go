package http

import (
	"encoding/json"
	"github.com/signmem/falcon-plus/modules/pingproxy/g"
	"github.com/signmem/falcon-plus/modules/pingproxy/tools"
	"net/http"
)



func configApiRoutes() {
	http.HandleFunc("/api/v1/pingcheck", hostPingCheck)
	http.HandleFunc("/_health_check", healthCheck)
}

func hostPingCheck(w http.ResponseWriter, r *http.Request) {
	var PingRequest  g.HttpPingRequest
	var PingResponse g.HttpPingResponse
	err := json.NewDecoder(r.Body).Decode(&PingRequest)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	PingResponse.PingStatus = tools.PingStatus(PingRequest.Ipaddr)
	PingResponse.Ipaddr = PingRequest.Ipaddr

	bs, err := json.Marshal(PingResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	g.Logger.Infof("pingcheck: ip:%s, sttus %b", PingResponse.Ipaddr, PingResponse.PingStatus)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bs)

}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok\n"))
}