package cache

import (
	"time"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
	"github.com/open-falcon/falcon-plus/modules/trend/proc"
)

const (
	CACHE_HISTORY_MAX_LEN          = 5
	GET_CACHE_HISTORY_LEN_INTERVAL = time.Duration(60) * time.Second
)

var (
	GaugeCacheHistory   *CacheHistory
	CounterCacheHistory *CacheHistory
	BigMapIndexArray    []string
	AllowMinKey         int64
	ExtractedKey        int64
	CacheHistoryIsFull  bool
)

func init() {
	GaugeCacheHistory = NewCacheHistory()
	CounterCacheHistory = NewCacheHistory()
	BigMapIndexArray = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	// AllowMinKey = time.Now().Unix()/g.TREND_INTERVALS + 1
	// ExtractedKey = AllowMinKey - 1
	CacheHistoryIsFull = false
}

func flushKey() {
	for {
		AllowMinKey = time.Now().Unix()/g.TREND_INTERVALS + 1
		ExtractedKey = AllowMinKey - 1
		time.Sleep(60 * time.Second)
	}
}


func checkHistoryLenCron() {
	for {
		time.Sleep(GET_CACHE_HISTORY_LEN_INTERVAL)
		checkHistoryLen()
	}
}

func checkHistoryLen() {
	if CounterCacheHistory == nil {
		CacheHistoryIsFull = true
		g.Logger.Warning("checkHistoryLen() counter cache history is nil")
		return
	} else {
		length := CounterCacheHistory.Len()
		if length >= CACHE_HISTORY_MAX_LEN {
			CacheHistoryIsFull = true
			g.Logger.Warningf("checkHistoryLen() length of counter cache history is larger than max length %d, length %d", CACHE_HISTORY_MAX_LEN, length)
			return
		}
	}
	if g.Config().Gauge {
		if GaugeCacheHistory == nil {
			CacheHistoryIsFull = true
			g.Logger.Warning("checkHistoryLen() gauge cache history is nil")
			return
		} else {
			length := GaugeCacheHistory.Len()
			if length >= CACHE_HISTORY_MAX_LEN {
				CacheHistoryIsFull = true
				g.Logger.Warningf("checkHistoryLen() length of guage cache history is larger than max length %d, length %d", CACHE_HISTORY_MAX_LEN, length)
				return
			}
		}
	}
	CacheHistoryIsFull = false
}

func Start() {
	go startExtractorCron()
	go checkHistoryLenCron()
	go flushKey()
}

func Push(items []*model.TrendItem) {
	if items == nil {
		return
	}

	count := len(items)
	if count == 0 {
		if g.Config().Debug {
			g.Logger.Debug("Cache.Push() push items length is zero.")
		}
		return
	}
	for _, item := range items {

		if g.Config().Debug {
			g.Logger.Debugf("Cache.Push() push item: %v", item)
		}

		if item == nil {
			if g.Config().Debug {
				g.Logger.Debug("Cache.Push() Push item is nil.")
			}
			continue
		}
		proc.RpcRecvCnt.Incr()
		key := item.Timestamp / g.TREND_INTERVALS

		if g.Config().Debug {
			g.Logger.Debugf("Cache.Push() key %v,  AllowMinKey %v, ExtractedKey %v", key, AllowMinKey, ExtractedKey)
		}

		if key < AllowMinKey || key <= ExtractedKey {
			if g.Config().Debug {
				g.Logger.Debugf("Cache.Push() skip because item key is [%d], and smaller than AllowMinKey [%d] or ExtractedKey [%d]", key, AllowMinKey, ExtractedKey)
			}
			proc.RpcKeyTooSmallCnt.Incr()
			continue
		}
		if key > ExtractedKey+2 {
			if g.Config().Debug {
				g.Logger.Debugf("Cache.Push() skip because item key is [%d], and larger than ExtractedKey + 2, ExtractedKey [%d]", key, ExtractedKey)
			}
			proc.RpcKeyTooBigCnt.Incr()
			continue
		}
		if CacheHistoryIsFull {
			if g.Config().Debug {
				g.Logger.Debug("Cache.Push() skip because CacheHistoryIsFull" )
			}
			proc.RcpCacheHistoryFullCnt.Incr()
			continue
		}
		pk := item.PrimaryKey()
		if item.DsType == "GAUGE" {
			if g.Config().Gauge {
				big_map, ok := GaugeCacheHistory.Get(key)
				if !ok {
					big_map = GaugeCacheHistory.Set(key, NewCacheBigMap(GuageNew, GuageUpdate, GuageTrendResult))

					if g.Config().Debug {
						g.Logger.Debugf("Cache.Push() !ok create Gauge cache big map, key:%d", key)
					}
				}

				if g.Config().Debug {
					g.Logger.Debugf("Cache.Push() ok update Gauge cache big map, key:%d", key)
				}
				big_map.M[pk[0:2]].Update(key, pk, item)
				proc.RpcRecvGuegeCnt.Incr()
			}
		} else if item.DsType == "COUNTER" {
			big_map, ok := CounterCacheHistory.Get(key)
			if !ok {
				big_map = CounterCacheHistory.Set(key, NewCacheBigMap(CounterNew, CounterUpdate, CounterTrendResult))

				if g.Config().Debug {
					g.Logger.Debugf("Cache.Push() !ok create counter cache big map, key:%d", key)
				}
			}
			if g.Config().Debug {
				g.Logger.Debugf("Cache.Push() ok update counter cache big map, key:%d", key)
			}
			//log.Debugf("update counter cache big map, key:%d", key)
			big_map.M[pk[0:2]].Update(key, pk, item)
			proc.RpcRecvCounterCnt.Incr()
		} else {
			g.Logger.Errorf("kafka string dstype is invalid, item: %v", item)
		}
	}
}
