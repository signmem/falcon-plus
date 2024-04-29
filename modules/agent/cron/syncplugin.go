package cron

import (
	"encoding/json"
	"github.com/zloylos/grsync"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/toolkits/file"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	STARTUP = false
)

func RsyncRequests() (g.RsyncResponse, error) {

	var rsyncHttpBody g.RsyncResponse

	httpurl := g.Config().RsyncAccess

	request, err := http.Get(httpurl)

	if err != nil {
		log.Println("RsyncRequests() http access error. ", err)
		return rsyncHttpBody, err
	}

	defer request.Body.Close()

	responseBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println("RsyncRequests() http response error. ", err)
		return rsyncHttpBody, err
	}
	if err := json.Unmarshal( responseBody, &rsyncHttpBody) ; err != nil {
		return rsyncHttpBody, err
	}

	return rsyncHttpBody, nil
}

func CheckIP(httpResponse g.RsyncResponse) bool {

	networkRange := httpResponse.Source
	_, subnet, _ := net.ParseCIDR(networkRange)
	myIPAddr := net.ParseIP(g.LocalIp)
	if subnet.Contains(myIPAddr) {
		return true
	}else {
		return false
	}


}

func GoRsyncPlugin(rsyncHttpResponse g.RsyncResponse ) bool {
	rsyncServers := rsyncHttpResponse.Server

	rand.Seed(int64(time.Now().UnixNano()))
	rIndex := rand.Intn(len(rsyncServers))

	rServer := rsyncServers[rIndex]
	rsyncDir := rsyncHttpResponse.SyncDest
	rsyncDelete := rsyncHttpResponse.SyncDelete
	rsyncSrc  := rsyncHttpResponse.SyncSrc
	rsyncPort := rsyncHttpResponse.SyncPort
	rsyncUser := g.Config().RsyncUser
	rsyncPwd  := g.Config().RsyncPwd

	_ = os.Setenv("RSYNC_PASSWORD", rsyncPwd)  // use to set rsync password into os env.

	if _, err := os.Stat(rsyncDir); os.IsNotExist(err) {
		g.UpdatePluginVersion("sync_dir_not_exists")
		// return false
	}

	var syncMethod string
	if strings.HasPrefix(rsyncSrc,"/") {
		syncMethod = "rsync://" + rsyncUser + "@" + rServer + ":" + rsyncPort +  rsyncSrc
	} else {
		syncMethod = "rsync://" + rsyncUser + "@" + rServer + ":" + rsyncPort + "/"  + rsyncSrc
	}

	var All []string
	All = append(All, "*")
	task := grsync.NewTask(
		syncMethod,
		rsyncDir,
		grsync.RsyncOptions{
			Dirs: true,
			Update: true,
			Include: All,
			Links: true,
			HardLinks: true,
			Owner: true,
			Group: true,
			Times: true,
			Delete:  rsyncDelete,
			Timeout:  600,
		},
	)

	go func() {
		for {
			_ = task.State()
			time.Sleep(time.Second)
		}
	}()

	if err := task.Run(); err != nil {
		log.Println("rsync faile :", err )
		return false
	}

	if g.UpdatePluginVersion(rsyncHttpResponse.Version) == true {
		return true
	}
	
	return false
}

func DoUpdateRpmNow(syncHttpResponse g.RsyncResponse, force bool) bool {
	if syncHttpResponse.RpmUpdate == true || force == true {
		if CheckIP(syncHttpResponse) == true || force == true {
			if syncHttpResponse.RpmVersion.El6 == "" || syncHttpResponse.RpmVersion.El7 == "" {
				log.Println("INFO: rpm update not needed")
				return false
			}
			if g.Config().Debug  {
				log.Printf("[INFO]: going to rpm update. ")
			}
			rpmUpdateStatus := RpmUpdate(syncHttpResponse)
			if rpmUpdateStatus == true {
				log.Println("INFO: rpm update success")
				return true
			} else {
				log.Println("ERROR: rpm update false")
				return false
			}
		}
	}
	log.Println("INFO: rpm update not needed")
	return false
}

func DoSyncPluginNow(rsyncHttpResponse g.RsyncResponse, syncDelete string, force bool) (msg string, status bool) {

	// if use api to access plugin update
	// must be do rsyncHttpResponse, err:= RsyncRequests() first

	if rsyncHttpResponse.Sync == true || force == true {
		if CheckIP(rsyncHttpResponse) == true || force == true {

			rsyncHttpResponse.SyncDest = g.Config().Plugin.Dir
			
			currentVersion := g.GetPluginVersion()
			re := regexp.MustCompile("@")
			if re.Match([]byte(currentVersion)) {
				versionList := strings.Split(currentVersion,"@")
				versionString := versionList[1]

				if  versionString !=  rsyncHttpResponse.Version  || force == true {

					if ! file.IsExist(rsyncHttpResponse.SyncDest) {
						os.Mkdir(rsyncHttpResponse.SyncDest, 0755)
					}

					err := unix.Access(rsyncHttpResponse.SyncDest, unix.W_OK)

					if err != nil {
						if g.ChownPluginDir(rsyncHttpResponse.SyncDest) == false {
							msg := "chown before sync error: " + rsyncHttpResponse.SyncDest
							log.Println(msg)
							g.UpdatePluginVersion("chown_plugin_dir_error_before_sync")
							return msg, false
						}
					}

					if g.UpdatePluginVersion(rsyncHttpResponse.Version) == true {

						if syncDelete == "enable" {
							rsyncHttpResponse.SyncDelete = true
						}

						status :=  GoRsyncPlugin(rsyncHttpResponse)

						if status == true {
							if g.ChownPluginDir(rsyncHttpResponse.SyncDest) == false {
								msg := "chown after sync error: " + rsyncHttpResponse.SyncDest
								log.Println(msg)
								g.UpdatePluginVersion("chown_plugin_dir_error_after_rsync")
								return  msg,false
							}
							msg := "SyncPlugin done and success"
							log.Println(msg)
							return msg,true
						} else {
							msg := "SyncPlugin() error, sync plugin error"
							g.UpdatePluginVersion("sync_plugin_error")
							log.Println(msg)
							return msg, false
						}
					}
				}
			}
		}
	}
	return "rsync plugin not needed", true
}

func SyncPlugin() {
	for {
		nowTime := g.GetNow()
		// log.Println( "now is ", nowTime)   // use for test
		// log.Println("job is ", g.JobTime )
		// var rsyncHttpResponse g.RsyncResponse

		if nowTime == g.JobTime || STARTUP == false {

			syncDelete := ""
			force := false

			rsyncHttpResponse, err:= RsyncRequests()

			if err != nil {
				msg := "plugin_http_access_error"
				g.UpdatePluginVersion(msg)
				time.Sleep(time.Second * 3600)
				continue
				// Do I need to send an alarm to pigeon ??
				// to be continue ......maybe ...???
			}


			_ = DoUpdateRpmNow(rsyncHttpResponse, force)

			_, status := DoSyncPluginNow(rsyncHttpResponse, syncDelete, force)
			if status == false {
				time.Sleep(time.Second * 120)
				continue
			}
			STARTUP = true
		}
		time.Sleep(time.Second * 60)
	}
}
