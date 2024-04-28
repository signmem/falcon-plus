package cron

import (
	"bytes"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var newRpmFile = "/tmp/falcon-agent-last-release.rpm"

func RpmUpdate(rsyncHttpResponse g.RsyncResponse) (bool) {

	// use to check rpm update

	rpmUpdate := rsyncHttpResponse.RpmUpdate

	if rpmUpdate == true {
		rpmInstallStatus := DoInstallRPM(rsyncHttpResponse)
		if rpmInstallStatus == false {
			return false
		} else {
			return true
		}

	} else {
		return false
	}

}

func DownLoadFile(rsyncHttpResponse g.RsyncResponse) bool {

	// notice : if use api update rpm
	// must be  do rsyncHttpResponse, err:= RsyncRequests() again.

	var rpmUrl string

	osVersion := g.CheckOSVersion()

	if _, err := os.Stat(newRpmFile) ; err == nil {
		os.Remove(newRpmFile)
	}

	if g.Config().Debug {
		log.Printf("[INFO] DownLoadFile() rpm install for os version: %s ", osVersion)
	}

	if osVersion == "6" {
		rpmUrl = rsyncHttpResponse.RpmVersion.El6
	} else if osVersion == "7" {
		rpmUrl = rsyncHttpResponse.RpmVersion.El7
	} else if osVersion == "8" {
		rpmUrl = rsyncHttpResponse.RpmVersion.El8
	}else if osVersion == "5" {
		rpmUrl = rsyncHttpResponse.RpmVersion.El5
	} else {
		log.Println("ERROR: DownLoadFile() os version not support.")
		return false
	}

	urlStatus := g.CheckUrlTcpHealth(rpmUrl)
	if urlStatus == false {
		log.Println( "ERROR: DownLoadFile() dns or firewall problem")
		return false
	}

	resp, err := http.Get(rpmUrl)
	if err != nil {
		log.Println("ERROR: DownLoadFile() rpm file http access faile. try wget ", rpmUrl )
		return false
	}
	defer resp.Body.Close()

	out, err := os.Create(newRpmFile)
	if err != nil {
		log.Println("ERROR: DownLoadFile() rpm file save to local faile.")
		return false
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("ERROR: rpm file save to local faile2.")
		return false
	}
	if g.Config().Debug {
		log.Printf("[INFO] rpm download into %s", newRpmFile)
	}
	return  true
}

func DoInstallRPM(rpmHttpBody g.RsyncResponse) bool {
	if g.Config().Debug {
		log.Println("[INFO] DoInstallRPM() goding to download rpm.")
	}
	downLoadStatus := DownLoadFile(rpmHttpBody)
	if downLoadStatus == false {
		return false
	}

	cmd := exec.Command("sudo", "rpm", "-Uvh", "--force", newRpmFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}