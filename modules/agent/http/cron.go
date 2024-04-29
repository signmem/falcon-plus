package http

import (
	"encoding/json"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"io/ioutil"
	"net/http"
	"strconv"
)

type TimeReset struct {
	Hour      int     `json:"hour"`
	Minute    int     `json:"minite"`
}

func configCronRoutes() {

	http.HandleFunc("/cron/reset", func(w http.ResponseWriter, r *http.Request) {
		g.GenCronTime()
		RenderDataJson(w, map[string]interface{}{
			"hour":    g.JobTime.Hour,
			"minite":  g.JobTime.Minite,
		})
	})

	http.HandleFunc("/cron/gettime", func(w http.ResponseWriter, r *http.Request) {
		RenderDataJson(w, map[string]interface{}{
			"hour":    g.JobTime.Hour,
			"minite":  g.JobTime.Minite,
		})
	})

	http.HandleFunc("/cron/settime", func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			http.Error(w, "body is blank", http.StatusBadRequest)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//	body := string(bs)
		var cronTime TimeReset

		if err := json.Unmarshal(bs, &cronTime) ; err == nil {

			if cronTime.Hour > 18 || cronTime.Hour < 10 {
				w.Write([]byte("error hour must in [ 10 - 18 ]\n"))
				return
			}
			if cronTime.Minute > 59 || cronTime.Minute < 0 {
				w.Write([]byte("error minite must in [ 0 - 59 ]\n"))
				return
			}
			if cronTime.Minute < 10 {
				minite := strconv.Itoa(cronTime.Minute)
				g.JobTime.Minite = "0" + minite
			} else {
				g.JobTime.Minite = strconv.Itoa(cronTime.Minute)
			}
			g.JobTime.Hour =   strconv.Itoa(cronTime.Hour)


			RenderDataJson(w, map[string]interface{}{
				"hour":    g.JobTime.Hour,
				"minite":  g.JobTime.Minite,
			})
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})
}
