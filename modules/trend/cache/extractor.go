package cache

import (
	"sort"
	"time"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
	nsema "github.com/toolkits/concurrent/semaphore"

	"github.com/toolkits/cron"
	ttime "github.com/toolkits/time"
)

const (
	//EXTRACTOR_DEFAULT_SLEEP_TIME       = time.Duration(1000) * time.Millisecond //default is 1000ms
	EXTRACTOR_DEFAULT_HISTORY_INTERVAL = 420 //默认下个时间段map创建超过7分钟，本时间段的数据才可以写入，确保本时间段数据cache完全
)

var (
	extractorCron     = cron.New()
	extractorCronSpec = "0 10 * * * ?"   // 理论上每小时第 10 分钟执行一次
)

func startExtractorCron() {
	extractorCron.AddFuncCC(extractorCronSpec, func() {
		if g.Config().DBLog {
			g.Logger.Debug("startExtractorCron() CRON going to start.")
		}
		start := time.Now().Unix()
		extract()
		end := time.Now().Unix()
		g.Logger.Infof("startExtractorCron CRON DONE, time used %ds, start at %s", end-start, ttime.FormatTs(start))
	}, 1)
	extractorCron.Start()
	select {}
}

func extract() {
	concurrent := g.Config().DB.Concurrent
	if concurrent < 1 {
		concurrent = 1
	}
	extractJob(CounterCacheHistory, concurrent, "counter")
	if g.Config().Gauge {
		extractJob(GaugeCacheHistory, concurrent, "gauge")
	}

}

func extractJob(history *CacheHistory, concurrent int, dstype string) {
	if history == nil {
		g.Logger.Warningf("extractJob() history map is nil, type: %s", dstype)
		return
	}
	length := history.Len()
	if g.Config().DBLog {
		g.Logger.Debugf("extractJob() history length is: %d", length)
	}
	if length <= 1 {
		g.Logger.Warningf("extractJob() no %s data need to extract, key length %d", dstype, length)
		return
	}
	keys := history.Keys()
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	if g.Config().DBLog {
		g.Logger.Debugf("extractJob() start extract %s data, keys %v ", dstype, keys)
	}
	sema := nsema.NewSemaphore(concurrent)
	for i := 0; i <= length-2; i++ {
		if g.Config().DBLog {
			g.Logger.Debugf("extractJob() start flush %s data loop, index [%d], " +
				"key [%d]", dstype, i, keys[i])
		}
		now := time.Now().Unix()
		now_key := now / g.TREND_INTERVALS
		if keys[i] >= now_key {
			if g.Config().DBLog {
				g.Logger.Debugf("SKIP because: extractJob() key is : %d now is %d",
					keys[i], now_key)
			}
			continue
		}
		big_map, exist := history.Get(keys[i])
		if !exist || big_map == nil {
			if g.Config().DBLog {
				g.Logger.Debugf("SKIP extractJob() current big map is not exist, " +
					"key: %d", keys[i])
			}
			continue
		}
		big_map_next, exist_next := history.Get(keys[i+1])
		if !exist_next || big_map_next == nil {
			if g.Config().DBLog {
				g.Logger.Debugf("SKIP extractJob() extract %s data, next map is not exist, " +
					"current key: %d, next key: %d", dstype, keys[i], keys[i+1])
			}
			continue
		}
		ts := now - big_map_next.GetCreateTime()
		if g.Config().DBLog {
			g.Logger.Debugf("SKIP extractJob() extract %s data, " +
				"next map created duration [%d], create time [%d], now [%d], " +
				"current key: %d, next key: %d", dstype, ts,
				big_map_next.GetCreateTime(), now, keys[i], keys[i+1])
		}
		if ts < EXTRACTOR_DEFAULT_HISTORY_INTERVAL {
			//下个时间段创建不到7分钟，不能确保上个时间段数据cache完全
			if g.Config().DBLog {
				g.Logger.Debugf("SKIP extractJob() extract %s data, next map " +
					"created duration is %d, smaller min interval %d, " +
					"current key: %d, next key: %d", dstype, ts,
					EXTRACTOR_DEFAULT_HISTORY_INTERVAL, keys[i], keys[i+1])
			}
			continue
		}
		ExtractedKey = keys[i]
		if g.Config().DBLog {
			g.Logger.Debugf("START extractJob() start flush %s data, key [%d], " +
				"start [%s]", dstype, keys[i], ttime.FormatTs(now))
		}
		for j := 0; j < 16; j++ {
			for k := 0; k < 16; k++ {
				big_key := BigMapIndexArray[j] + BigMapIndexArray[k]
				item_map, exist := big_map.Get(big_key)
				if !exist {
					if g.Config().DBLog {
						g.Logger.Debugf("SKIP2 extractJob() %s big map data is not exist, " +
							"big map key [%d]", dstype, big_key)
					}
					continue
				} else if item_map == nil {
					if g.Config().DBLog {
						g.Logger.Debugf("SKIP2 extractJob() %s item map is nil, " +
							"big map key [%d]", dstype, big_key)
					}
					big_map.Delete(big_key)
					continue
				}
				// log.Debugf("[DEBUG] START extractJob() start flush %s data, " +
				// 	"big map key [%s]", dstype, big_key)
				//  每次出现日志 big_key = [00] - [ff] 意义不大 
				sema.Acquire()
				go func(m *ItemMap, key int64) {
					defer sema.Release()
					if m != nil {
						m.Flush(key)
						m.DeleteAll()
						return
					}
				}(item_map, keys[i])
				big_map.Delete(big_key)
				//time.Sleep(EXTRACTOR_DEFAULT_SLEEP_TIME)
			}
		}
		history.Delete(keys[i])
		g.Logger.Debugf("extractJob() Delete %s data, key [%d], finsh [%s]", dstype, keys[i], ttime.FormatTs(time.Now().Unix()))
	}
}
