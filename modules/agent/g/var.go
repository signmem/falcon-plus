package g

import (
	"bytes"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/toolkits/slice"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var Root string

func InitRootDir() {
	var err error
	Root, err = os.Getwd()
	if err != nil {
		log.Fatalln("getwd fail:", err)
	}
}

var LocalIp string

func ShellOut(command string)  {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Println(err)
	}else{
		log.Println("[INFO] ", command )
		log.Println("[INFO] ", stdout.String() )
		if stderr.String() != "" {
			log.Println("[ERROR] ", command)
			log.Println("[ERROR] ", stderr.String() )
		}
	}
}

func InitLocalIp() {
	if Config().Heartbeat.Enabled {

		LocalIp = ""

		if runtime.GOOS != "windows" && os.Getuid() == 0 {

			// try get ip from ssh first.
			// sshd_config must read by root
			ipaddr, err := getSSHIpaddr()

			if err != nil {
				log.Println(err)
			} else {
				LocalIp = ipaddr
				log.Println("get ip addr from ssh config, local ip is now:", LocalIp)
			}
		} else {
			log.Println("[WRN]: not running by root user.")
		}


		for LocalIp == "" {
			ipaddr,  err := getLinuxIPaddress()
			if err != nil || ipaddr == ""  {
				systemIP := "/sbin/ip addr"
				systemRoute := "/sbin/ip route"
				log.Println("InitLocalIp() error ", err)

				ShellOut(systemIP)
				ShellOut(systemRoute)
			}
			LocalIp = ipaddr
			log.Println("Local IP is now : ", LocalIp)
			if LocalIp != "" {
				break
			}else {
				time.Sleep( time.Second * 30 )
			}
		}
	} else {
		log.Println("hearbeat is not enabled, can't get localip")
	}
}

var (
	HbsClient *SingleConnRpcClient
)

func InitRpcClients() {
	if Config().Heartbeat.Enabled {
		HbsClient = &SingleConnRpcClient{
			RpcServer: Config().Heartbeat.Addr,
			Timeout:   time.Duration(Config().Heartbeat.Timeout) * time.Millisecond,
		}
	}
}

func SendToTransfer(metrics []*model.MetricValue) {
	if len(metrics) == 0 {
		return
	}

	debug := Config().Debug

	dt := Config().DefaultTags

	if len(dt) > 0 {
		var buf bytes.Buffer
		default_tags_list := []string{}
		for k, v := range dt {
			buf.Reset()
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(v)
			default_tags_list = append(default_tags_list, buf.String())
		}
		default_tags := strings.Join(default_tags_list, ",")

		for i, x := range metrics {
			buf.Reset()

			if x.Tags == "" {
				metrics[i].Tags = default_tags
			} else {
				buf.WriteString(metrics[i].Tags)
				buf.WriteString(",")
				buf.WriteString(default_tags)
				metrics[i].Tags = buf.String()
			}
		}
	}

	// add deploy pool  tags for every metric ( edit by terry tsang )

	pool := POOLNAME
	var default_tags string
	if VirtualMachine {
		default_tags = "pool=" + pool + ",family=virtual"
	} else {
		default_tags = "pool=" + pool
	}

	//var buf bytes.Buffer
	for i, x := range metrics {
		if x.Metric == "agent.alive" {
			continue
		}

		if x.Tags == "" {
			metrics[i].Tags = default_tags
		} else {
			oldTags := x.Tags
			if ! strings.Contains(oldTags, default_tags) {
				newTags := strings.Join([]string{oldTags,default_tags},",")
				x.Tags = newTags
			}
		}
	}


	if debug {
		log.Println("##########################  debug metric start ############################")
		for l, _ := range metrics {
			log.Printf("debug metric name: %v\n", metrics[l] )
		}
		log.Println("##########################  debug metric end ############################")
		log.Printf("===> <Total=%d> %v\n", len(metrics), metrics[0])
	}

	metricCount := ReportMetricCounts()
	totalMetricCount := metricCount + int64(len(metrics))
	SetMetricCount(totalMetricCount)

	var resp model.TransferResponse
	SendMetrics(metrics, &resp)

	if debug {
		log.Println("<===", &resp)
	}
}

var (
	reportUrls     map[string]string
	reportUrlsLock = new(sync.RWMutex)
)

func ReportUrls() map[string]string {
	reportUrlsLock.RLock()
	defer reportUrlsLock.RUnlock()
	return reportUrls
}

func SetReportUrls(urls map[string]string) {
	reportUrlsLock.RLock()
	defer reportUrlsLock.RUnlock()
	reportUrls = urls
}

var (
	reportMetricCounts int64
	metricCountLock = new(sync.RWMutex)
)

func ReportMetricCounts() int64 {
	// user to get var reportMetricCounts <- []int64
	metricCountLock.RLock()
	defer metricCountLock.RUnlock()
	return reportMetricCounts
}

func SetMetricCount(metricCount int64) {
	// use to set var reportMetricCounts <- []int64
	metricCountLock.Lock()
	defer metricCountLock.Unlock()
	reportMetricCounts = metricCount
}


var (
	reportPorts     []int64
	reportPortsLock = new(sync.RWMutex)
)

func ReportPorts() []int64 {
	reportPortsLock.RLock()
	defer reportPortsLock.RUnlock()
	return reportPorts
}

func SetReportPorts(ports []int64) {
	reportPortsLock.Lock()
	defer reportPortsLock.Unlock()
	reportPorts = ports
}

var (
	duPaths     []string
	duPathsLock = new(sync.RWMutex)
)

func DuPaths() []string {
	duPathsLock.RLock()
	defer duPathsLock.RUnlock()
	return duPaths
}

func SetDuPaths(paths []string) {
	duPathsLock.Lock()
	defer duPathsLock.Unlock()
	duPaths = paths
}

var (
	// tags => {1=>name, 2=>cmdline}
	// e.g. 'name=falcon-agent'=>{1=>falcon-agent}
	// e.g. 'cmdline=xx'=>{2=>xx}
	reportProcs     map[string]map[int]string
	reportProcsLock = new(sync.RWMutex)
)

func ReportProcs() map[string]map[int]string {
	reportProcsLock.RLock()
	defer reportProcsLock.RUnlock()
	return reportProcs
}

func SetReportProcs(procs map[string]map[int]string) {
	reportProcsLock.Lock()
	defer reportProcsLock.Unlock()
	reportProcs = procs
}

var (
	ips     []string
	ipsLock = new(sync.Mutex)
)

func TrustableIps() []string {
	ipsLock.Lock()
	defer ipsLock.Unlock()
	return ips
}

func SetTrustableIps(ipStr string) {
	arr := strings.Split(ipStr, ",")
	ipsLock.Lock()
	defer ipsLock.Unlock()
	ips = arr
}

func IsTrustable(remoteAddr string) bool {
	ip := remoteAddr
	idx := strings.LastIndex(remoteAddr, ":")
	if idx > 0 {
		ip = remoteAddr[0:idx]
	}

	if ip == "127.0.0.1" {
		return true
	}

	return slice.ContainsString(TrustableIps(), ip)
}
