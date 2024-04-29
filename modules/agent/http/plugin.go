package http

import (
	"fmt"
	"github.com/signmem/falcon-plus/modules/agent/cron"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"net/http"
)

func configPluginRoutes() {
	http.HandleFunc("/plugin/update", func(w http.ResponseWriter, r *http.Request) {
		if !g.Config().Plugin.Enabled {
			w.Write([]byte("plugin not enabled\n"))
			return
		}
		syncDelete := ""
		force := true
		rsyncHttpResponse, err:= cron.RsyncRequests()
		if err != nil {
			w.Write([]byte("[ERROR]: rsync http access for json info faile.\n"))
			return
		}
		msg, status := cron.DoSyncPluginNow(rsyncHttpResponse, syncDelete, force)

		if status == false {
			w.Write([]byte(fmt.Sprintf(msg + "\n")))
			return
		}

		w.Write([]byte("success\n"))
	})

	http.HandleFunc("/plugin/reset", func(w http.ResponseWriter, r *http.Request) {
		if !g.Config().Plugin.Enabled {
			w.Write([]byte("plugin not enabled\n"))
			return
		}
		pluginDir := g.Config().Plugin.Dir
		versionFile := pluginDir + "/version"
		g.ForceResetVersion(versionFile)

		syncDelete := "enable"
		force := true
		rsyncHttpResponse, err:= cron.RsyncRequests()
		if err != nil {
			w.Write([]byte("[ERROR]: rsync http access for json info faile.\n"))
			return
		}
		msg, status := cron.DoSyncPluginNow(rsyncHttpResponse, syncDelete, force)
		if status == false {
			w.Write([]byte(fmt.Sprintf(msg + "\n")))
			return
		}

		w.Write([]byte("success\n"))
	})

	http.HandleFunc("/plugin/version", func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte(g.GetPluginVersion() + "\n"))
	})
}
