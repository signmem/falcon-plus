package http

import (
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"net/http"
)

func configHealthRoutes() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		falconVersion := g.VERSION
		pluginVersion := g.GetPluginVersion()

		RenderDataJson(w, map[string]interface{}{
			"falcon-agent":    falconVersion,
			"falcon-plugin":  pluginVersion,
		})
	})
}
