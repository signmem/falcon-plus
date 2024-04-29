package http

import (
	"fmt"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/sender"
	"net/http"
	"strings"
)

func configDebugHttpRoutes() {
	// conn pools
	http.HandleFunc("/debug/connpool/", func(w http.ResponseWriter, r *http.Request) {
		urlParam := r.URL.Path[len("/debug/connpool/"):]
		args := strings.Split(urlParam, "/")

		argsLen := len(args)
		if argsLen < 1 {
			w.Write([]byte(fmt.Sprintf("bad args\n")))
			return
		}

		var result string
		receiver := args[0]
		switch receiver {
		case "trend":
			result = strings.Join(sender.TrendConnPools.Proc(), "\n")
		default:
			result = fmt.Sprintf("bad args, module not exist\n")
		}
		w.Write([]byte(result))
	})
}
