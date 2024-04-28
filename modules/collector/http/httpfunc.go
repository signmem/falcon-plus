package http

import (
	_ "encoding/json"
	_ "io/ioutil"
	_ "log"
	"net/http"
)


func healthCheck() {
	http.HandleFunc("/_health_check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
}
