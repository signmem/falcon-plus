// Package index 实现 Open-Falcon Graph 模块索引异步增量更新逻辑
// 作用：消费未索引缓存队列，异步将指标元数据批量落地到MySQL，分担实时上报压力
package index

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"time"

	nsema "github.com/toolkits/concurrent/semaphore"
	ntime "github.com/toolkits/time"

	"github.com/signmem/falcon-plus/modules/graph/g"
	proc "github.com/signmem/falcon-plus/modules/graph/proc"
)

const (
	// IndexUpdateIncrTaskSleepInterval 增量索引更新任务轮询间隔
	// 每隔1秒执行一次增量消费逻辑
	// 增量更新间隔时间, 默认1s
	IndexUpdateIncrTaskSleepInterval = time.Duration(1) * time.Second 
)

var (
	// 索引增量更新时操作mysql的并发控制
	// semaUpdateIndexIncr MySQL写入并发信号量
	// 限制增量更新同时操作数据库的协程数，避免DB连接打满、压力过载
	semaUpdateIndexIncr = nsema.NewSemaphore(2) 
)

// 启动索引的 异步、增量更新 任务, 每隔一定时间，刷新cache中的数据到数据库中
func StartIndexUpdateIncrTask() {

	// StartIndexUpdateIncrTask 启动索引异步增量更新常驻任务
	// 独立后台循环任务：定时从(未索引缓存)拉取数据，异步刷入MySQL

	for {
		// 按配置间隔休眠，控制轮询频率
		time.Sleep(IndexUpdateIncrTaskSleepInterval)

		startTs := time.Now().Unix()
		// 执行单次增量更新逻辑，返回本次处理的数据条数
		cnt := updateIndexIncr()
		endTs := time.Now().Unix()
		// statistics
		// 上报运行监控指标：处理条数、执行次数、开始时间、耗时
		proc.IndexUpdateIncrCnt.SetCnt(int64(cnt))
		proc.IndexUpdateIncr.Incr()
		proc.IndexUpdateIncr.PutOther("lastStartTs", ntime.FormatTs(startTs))
		proc.IndexUpdateIncr.PutOther("lastTimeConsumingInSec", endTs-startTs)
	}
}

// 进行一次增量更新
func updateIndexIncr() int {
	// updateIndexIncr 执行单次索引增量更新
	// 消费 unIndexedItemCache 未索引缓存队列，异步入库MySQL
	// 返回值：本次成功发起入库的指标总条数

	ret := 0

	// 未索引缓存为空，直接返回0，无需处理
	if unIndexedItemCache == nil || unIndexedItemCache.Size() <= 0 {
		return ret
	}

	// 调试模式：打印未索引缓存对象信息
	if g.Config().Debug {
		log.Printf("[DEBUG] unIndexedItemCache is %v", unIndexedItemCache)
	}

	// 获取数据库连接
	dbConn, err := g.GetDbConn("UpdateIndexIncrTask")
	if err != nil {
		log.Error("[ERROR] get dbConn fail", err)
		return ret
	}

	// 获取未索引缓存中所有key
	keys := unIndexedItemCache.Keys()

	// 遍历所有待索引缓存项
	for _, key := range keys {
		// 取出缓存项并立即从(未索引队列)中移除，防止重复消费
		icitem := unIndexedItemCache.Get(key)
		unIndexedItemCache.Remove(key)
		if icitem != nil {
			// 并发更新mysql

			// 占用DB并发信号量，控制MySQL写入并发
			semaUpdateIndexIncr.Acquire()
			// 启动协程异步入库，不阻塞主循环
			go func(key string, icitem *IndexCacheItem, dbConn *sql.DB) {
				// 协程退出必释放信号量，避免死锁
				defer semaUpdateIndexIncr.Release()
				// 执行单条索引入库SQL操作
				err := updateIndexFromOneItem(icitem.Item, dbConn)
				if err != nil {
					// 入库失败：累加错误计数
					proc.IndexUpdateIncrErrorCnt.Incr()
				} else {
					// 入库成功：将数据迁移至(已索引缓存)IndexedItemCache
					IndexedItemCache.Put(key, icitem)
				}
			}(key, icitem.(*IndexCacheItem), dbConn)
			ret++
		}
	}

	return ret
}
