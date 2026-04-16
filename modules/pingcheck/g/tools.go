package g

import (
	"fmt"
	"net"
	"os"
	"time"
)


func MonitorAgentPeriod() {

	// 只针对 agent 设定降级 

	for {

		agentAlarmHost := FalconAgentLru.GetLength(0)

		if 	agentAlarmHost >= Config().Degrade.AlarmLimit {

			SkipAgentAlarm = true
			Logger.Debugf("[AGENT 报警降级开始] SkipAgentAlarm 开启为 true")
			time.Sleep(time.Duration( Config().Degrade.FrozenTime ) * time.Minute )
			SkipAgentAlarm = false
			Logger.Debugf("[AGENT 报警报警结束] SkipAgentAlarm 开启为 false")

		} else {
			time.Sleep(10 * time.Second)
		}

		// terry.zeng

	}
}


func LruMaintain() {
	FalconAgentLru.Init()  // use to init Lru [timenow] = []string
	FalconPingLru.Init()   //
	t1 := time.Now().Add( - time.Duration( Config().Degrade.Period ) * time.Minute )
	timeStamp := t1.Format("2006-01-02 15:04")
	FalconAgentLru.Delete(timeStamp)
	FalconPingLru.Delete(timeStamp)
}

func MonitorPingPeriod() {

	// 只针对 ping 设定降级

	for {
		agentAlarmHost := FalconPingLru.GetLength(0)

        if agentAlarmHost >= Config().Degrade.AlarmLimit {

			SkipPingAlarm = true
			Logger.Debugf("[PING 报警降级开始] SkipPingAlarm 开启为 true")
			time.Sleep(time.Duration( Config().Degrade.FrozenTime ) * time.Minute )
			SkipPingAlarm = false
			Logger.Debugf("[PING 报警降级结束] SkipPingAlarm 开启为 false")

		} else {
			time.Sleep(10 * time.Second)
		}
	}
}

func GetIP() string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return "0.0.0.0"
}

func (l *Lru) GetLength(line int) (length int) {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	var totalLine int
	lineNow := len(l.TotalLru)

	if  lineNow == 0 {

		return 0

	} else {
		if line > lineNow {
			totalLine = line
		} else {
			totalLine = lineNow
		}
	}

	num := 0;
	for num != totalLine {
		t1 := time.Now().Add( - time.Duration(num) * time.Minute )
		timeNow := t1.Format("2006-01-02 15:04")
		length += len(l.TotalLru[timeNow])
		num += 1
	}
	return
}



func (l *Lru) Init() {
	l.Mu.Lock()
	l.Mu.Unlock()
	t1 := time.Now()
	timeNow := t1.Format("2006-01-02 15:04")

	hostList :=  make([]string,0)
	if len(l.TotalLru[timeNow]) == 0 {
		l.TotalLru[timeNow] = hostList
	}
}

func (l *Lru) GetLruDetail(line int) string {

	// line = 0 then return 99 limit lines info
	// line = num then retuen num line.

	l.Mu.Lock()
	defer l.Mu.Unlock()
	var str string
	var totalLine int

	lineNow := len(l.TotalLru)

	if  lineNow == 0 {
		t1 := time.Now().Format("2006-01-02 15:04")
		str = fmt.Sprintf("[time: %s, line: %d ]", t1, lineNow)
		return str
	} else {
		if line > lineNow {
			totalLine = line
		} else {
			totalLine = lineNow
		}
	}

	num := 0;
	for num != totalLine {
		t1 := time.Now().Add( - time.Duration(num) * time.Minute )
		timeNow := t1.Format("2006-01-02 15:04")
		str += fmt.Sprintf("[time:%s, info: %v ] ", timeNow, l.TotalLru[timeNow])
		num += 1
	}
	return str
}

func (l *Lru) TimeMatch(timeNow string) bool {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	for timeStamp, _ := range l.TotalLru {
		if timeStamp == timeNow {
			return true
		}
	}
	return false
}

func (l *Lru) NewHost(timeNow string, host string) () {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	newCache := make(map[string][]string)
	var hostList []string
	hostList = append(hostList, host)
	newCache[timeNow] = hostList
	id := 1
	for id < Config().Degrade.Period {
		t1 := time.Now().Add( - time.Duration(id) * time.Minute )
		timeStamp := t1.Format("2006-01-02 15:04")

		newCache[timeStamp] = l.TotalLru[timeStamp]
		id += 1
	}
	l.TotalLru = newCache
}

func (l *Lru) Append(hostInfo string) {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	t := time.Now()
	timeNow := t.Format("2006-01-02 15:04")

	hosts := l.TotalLru[timeNow]
	hosts = append(hosts, hostInfo)
	l.TotalLru[timeNow] = hosts

}

func (l *Lru) Delete(timeNow string) {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	delete(l.TotalLru, timeNow)
}


func SliceContains(str string, s []string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func IntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
