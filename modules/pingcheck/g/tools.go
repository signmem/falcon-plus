package g

import (
	"net"
	"os"
	"time"
)

func PutData(totalCache map[int]LruCache, newCache LruCache) (reNewCache map[int]LruCache) {

	// 用于把一分钟内数据放入 reNewCache (时间周期)

	var numcounter int
	reNewCache = make(map[int]LruCache, Config().Degrade.Period)

	if  len(totalCache) <= Config().Degrade.Period {
		numcounter = len(totalCache)
	} else {
		numcounter = Config().Degrade.Period
	}

	for i := 0 ; i < numcounter  ; i++ {
		newid := i + 1
		reNewCache[newid] =  totalCache[i]
	}
	reNewCache[0] = newCache
	return reNewCache
}

func GetHostCount(totalCache map[int]LruCache)  ( totalCount int ) {

	// 用于计算时间周期内一共有多少主机经过告警

	totalCount = 0
	if len(totalCache) == 0 {
		return totalCount
	}

	t := time.Now().Unix() - int64(Config().Degrade.FrozenTime * 60)

	for _, info := range totalCache {
		if info.Timestamp >= t {
			totalCount = info.Len() + totalCount
		}
	}
	return totalCount
}


func MonitorPeriod() {

	// use to set || unset  period

	for {

		agentAlarmHost := GetHostCount(TotalLru)

		if 	agentAlarmHost >= Config().Degrade.AlarmLimit {

			SkipAlarm = true

			time.Sleep(time.Duration( Config().Degrade.FrozenTime ) * time.Second * 60 )

			// 如果希望在降级期结束后统一清理主机列表，则打开清空 g.TotalLru 方法
			// g.TotalLru = make(map[int]g.LruCache, g.Config().Degrade.Period)
			SkipAlarm = false

		} else {
			time.Sleep(60 * time.Second)
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
