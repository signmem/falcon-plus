package http

import (
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/signmem/nux"
	"github.com/signmem/sys"
	"net/http"
)

func configKernelRoutes() {
	http.HandleFunc("/proc/kernel/hostname", func(w http.ResponseWriter, r *http.Request) {
		AutoRender(w, g.HostName, nil)
	})

	http.HandleFunc("/proc/kernel/maxproc", func(w http.ResponseWriter, r *http.Request) {
		data, err := nux.KernelMaxProc()
		AutoRender(w, data, err)
	})

	http.HandleFunc("/proc/kernel/maxfiles", func(w http.ResponseWriter, r *http.Request) {
		data, err := nux.KernelMaxFiles()
		AutoRender(w, data, err)
	})

	http.HandleFunc("/proc/kernel/version", func(w http.ResponseWriter, r *http.Request) {
		data, err := sys.CmdOutNoLn("uname", "-r")
		AutoRender(w, data, err)
	})

}
