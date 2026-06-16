package index

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"sync"

	nsema "github.com/toolkits/concurrent/semaphore"
	ntime "github.com/toolkits/time"

	cmodel "github.com/signmem/falcon-plus/common/model"
	cutils "github.com/signmem/falcon-plus/common/utils"
	"github.com/signmem/falcon-plus/modules/graph/g"
	proc "github.com/signmem/falcon-plus/modules/graph/proc"
)

const (
	// DefaultUpdateStepInSec 默认全量索引更新时间窗口：2天(单位秒)
	// 只处理 最近2天内产生的指标数据，早于该时间的数据直接清理
	// 更新步长,一定不能大于删除步长. 两天内的数据,都可以用来建立索引
	DefaultUpdateStepInSec     = 2 * 24 * 3600 

	// ConcurrentOfUpdateIndexAll 全量索引更新任务最大并发数（全局只允许1个全量任务运行）
	ConcurrentOfUpdateIndexAll = 1

	// terry 20260611
	// 新增：索引写库防抖窗口 单位：秒，endpoint + tag 每小时只可以更新一次
	IndexDbThrottleInterval = 3600
	// 定时清理间隔 30分钟(秒)
	ThrottleMapCleanInterval   = 1800   
	// 节流Key过期时间 2 天
	ThrottleKeyExpireSec       = 2 * 24 * 3600 
)

var (
	// semaIndexUpdateAllTask 全量索引更新任务总控信号量：限制同时只能有1个全量更新任务执行
	//全量同步任务 并发控制器
	semaIndexUpdateAllTask = nsema.NewSemaphore(ConcurrentOfUpdateIndexAll)

	// semaIndexUpdateAll MySQL写入并发信号量：单任务内最多4个协程同时操作数据库
	// 索引全量更新时的mysql操作并发控制
	semaIndexUpdateAll     = nsema.NewSemaphore(4)

	// terry 20260611
	// 新增：并发安全Map，key=指标唯一标识，value=最后一次写库时间戳(秒)
	indexThrottleMap sync.Map
)


func init() {
	// 启动后台定时清理协程，防止sync.Map内存泄漏
	go func() {
		ticker := time.NewTicker(time.Duration(ThrottleMapCleanInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().Unix()
			// 遍历所有节流Key
			indexThrottleMap.Range(func(key, val interface{}) bool {
				lastWriteTs := val.(int64)
				// 超过过期时间，删除Key
				if now-lastWriteTs > ThrottleKeyExpireSec {
					indexThrottleMap.Delete(key)
				}
				return true
			})
		}
	}()
}



func allowUpdateIndexDb(key string) bool {
	// allowUpdateIndexDb 节流判断：同一指标 在限制时间内 内只允许写库一次
	// key: endpoint+counter 拼接的唯一标识
	// 返回 true=允许写库，false=跳过

	now := time.Now().Unix()

	// 查询最后写库时间
	lastTs, exist := indexThrottleMap.Load(key)
	if !exist {
		// 无记录：首次写库，记录当前时间并放行
		indexThrottleMap.Store(key, now)
		return true
	}

	// 计算时间差
	lastWriteTs := lastTs.(int64)
	if now-lastWriteTs >= IndexDbThrottleInterval {
		// 超过防抖窗口：更新时间并放行
		indexThrottleMap.Store(key, now)
		return true
	}

	// 窗口内重复请求：直接跳过写库
	return false
}



func GetConcurrentOfUpdateIndexAll() int {
	// GetConcurrentOfUpdateIndexAll 获取当前正在运行的 全量索引更新任务 数量
	// 返回值：当前活跃任务数
	// 索引全量更新的当前并行数
	log.Debugf("[DEBUG] going to Flush GetConcurrentOfUpdateIndexAll()")

	// 总许可数 - 剩余许可数 = 已占用(正在执行)的任务数
	return ConcurrentOfUpdateIndexAll - semaIndexUpdateAllTask.AvailablePermits()
}


func UpdateIndexAllByDefaultStep() {
	// UpdateIndexAllByDefaultStep 使用默认时间窗口(2天)执行全量索引更新
	// 对外暴露的快捷入口
	// 索引的全量更新
	log.Debugf("[DEBUG] going to Flush UpdateIndexAllByDefaultStep()")
	UpdateIndexAll(DefaultUpdateStepInSec)
}

func UpdateIndexAll(updateStepInSec int64) {
	// UpdateIndexAll 全量更新索引入口函数
	// updateStepInSec: 时间窗口(秒)，只处理该时间范围内的指标数据

	// 减少任务积压,但高并发时可能无效(AvailablePermits不是线程安全的)
	if semaIndexUpdateAllTask.AvailablePermits() <= 0 {
		log.Println("[DEBUG] updateIndexAll, concurrent not available")
		return
	}

	// 占用全量任务信号量，函数退出时自动释放
	semaIndexUpdateAllTask.Acquire()
	defer semaIndexUpdateAllTask.Release()

	// 记录任务开始时间
	startTs := time.Now().Unix()
	// 执行核心全量更新逻辑，返回本次处理的指标条数
	cnt := updateIndexAll(updateStepInSec)
	endTs := time.Now().Unix()
	// 打印任务耗时、时间窗口、处理条数日志
	log.Printf("[DEBUG] UpdateIndexAll, lastStartTs %s, updateStepInSec %d, lastTimeConsumingInSec %d\n",
		ntime.FormatTs(startTs), updateStepInSec, endTs-startTs)

	// statistics
	// 上报监控打点：更新计数、开始时间、耗时、处理条数等运行状态
	proc.IndexUpdateAllCnt.SetCnt(int64(cnt))
	proc.IndexUpdateAll.Incr()
	proc.IndexUpdateAll.PutOther("lastStartTs", ntime.FormatTs(startTs))
	proc.IndexUpdateAll.PutOther("updateStepInSec", updateStepInSec)
	proc.IndexUpdateAll.PutOther("lastTimeConsumingInSec", endTs-startTs)
	proc.IndexUpdateAll.PutOther("updateCnt", cnt)
}

func UpdateIndexOne(endpoint string, metric string, tags map[string]string, dstype string, step int) error {
	// 更新一条监控数据对应的索引. 用于手动添加索引,一般情况下不会使用
	// UpdateIndexOne 单条指标增量更新索引（单条数据索引更新入口）
	// endpoint: 监控端点名称
	// metric: 指标名称
	// tags: 指标标签集合
	// dstype: 数据类型（GAUGE/COUNTER等）
	// step: 数据采集步长
	// 返回值：执行错误

	itemDemo := &cmodel.GraphItem{
		Endpoint: endpoint,
		Metric:   metric,
		Tags:     tags,
		DsType:   dstype,
		Step:     step,
	}
	md5 := itemDemo.Checksum()
	uuid := itemDemo.UUID()

	// 从本地缓存查询该指标的索引缓存项
	cached := IndexedItemCache.Get(md5)
	if cached == nil {
		return fmt.Errorf("not found")
	}

	// 类型断言为索引缓存对象
	icitem := cached.(*IndexCacheItem)

	// 校验UUID：判断指标类型/步长是否发生变更，不一致则返回错误
	if icitem.UUID != uuid {
		return fmt.Errorf("bad type or step")
	}
	// 取出原始监控指标对象
	gitem := icitem.Item

	// terry 20250611
	// ========== 单独给实时上报加节流 ==========
	counter := gitem.Metric
	if len(gitem.Tags) > 0 {
		counter = fmt.Sprintf("%s/%s", counter, cutils.SortedTags(gitem.Tags))
	}
	throttleKey := fmt.Sprintf("%s||%s", gitem.Endpoint, counter)
	if !allowUpdateIndexDb(throttleKey) {
		proc.CounterCacheDropCnt.Incr()
		return nil
	}
	// ========================================


	// 获取数据库连接
	dbConn, err := g.GetDbConn("UpdateIndexIncrTask")
	if err != nil {
		log.Println("[ERROR] make dbConn fail", err)
		return err
	}
	// 执行单条数据索引入库
	return updateIndexFromOneItem(gitem, dbConn)
}


func updateIndexAll(updateStepInSec int64) int {

	// updateIndexAll 全量索引更新核心逻辑
	// updateStepInSec: 时间窗口(秒)，过滤过期数据
	// 返回值：本次成功发起入库的指标总条数

	var ret int = 0

	// 缓存为空，直接返回0
	if IndexedItemCache == nil || IndexedItemCache.Size() <= 0 {
		return ret
	}

	// 获取数据库连接
	dbConn, err := g.GetDbConn("UpdateIndexIncrTask")
	if err != nil {
		log.Println("[ERROR] make dbConn fail", err)
		return ret
	}

	// lastTs for update index
	// 计算时间阈值：早于该时间的指标判定为过期数据
	ts := time.Now().Unix()
	lastTs := ts - updateStepInSec

	// 遍历缓存中所有指标的key
	keys := IndexedItemCache.Keys()
	for _, key := range keys {
		icitem := IndexedItemCache.Get(key)
		if icitem == nil {
			continue
		}

		// 调试模式打印缓存key
		if g.Config().Debug {
			log.Printf("[DEBUG] key from IndexedItemCache: %s", key)
		}

		// 取出原始指标对象
		gitem := icitem.(*IndexCacheItem).Item

		// 过滤过期数据：指标时间早于时间窗口阈值，从缓存中删除并跳过入库
		//缓存中的数据太旧了,不能用于索引的全量更新
		if gitem.Timestamp < lastTs { 

			if g.Config().Debug {
				log.Debugf("[DEBUG] remove from IndexedItemCache: metric %s, endpoint %s, time %d ",
					gitem.Metric, gitem.Endpoint, gitem.Timestamp)
			}
			// 清理过期缓存
			IndexedItemCache.Remove(key) //在这里做个删除,有点恶心
			continue
		}
		// 并发写mysql
		// 占用MySQL并发信号量，启动协程异步入库（控制DB并发）
		semaIndexUpdateAll.Acquire()
		go func(gitem *cmodel.GraphItem, dbConn *sql.DB) {
			defer semaIndexUpdateAll.Release()
			err := updateIndexFromOneItem(gitem, dbConn)
			if g.Config().Debug {
				log.Debugf("[DEBUG] update from semaIndexUpdateAll: metric %s, endpoint %s, time %d ",
					gitem.Metric, gitem.Endpoint, gitem.Timestamp)
			}
			if err != nil {
				// 入库异常，上报错误计数
				proc.IndexUpdateAllErrorCnt.Incr()
			}
		}(gitem, dbConn)

		ret++
	}

	return ret
}

// 根据item,更新mysql
func updateIndexFromOneItem(item *cmodel.GraphItem, conn *sql.DB) error {

	// updateIndexFromOneItem 单条指标写入MySQL索引表（核心DB操作函数）
	// 依次操作三张表：endpoint、tag_endpoint、endpoint_counter
	// item: 原始监控指标对象
	// conn: 数据库连接
	// 返回值：DB操作错误

	if item == nil {
		return nil
	}

	// ===================== 新增：节流防抖逻辑 START =====================
	// 1. 拼接和数据库一致的 counter（包含排序tags）
	counter := item.Metric
	if len(item.Tags) > 0 {
		counter = fmt.Sprintf("%s/%s", counter, cutils.SortedTags(item.Tags))
	}
	// 2. 拼接全局唯一Key: endpoint + counter
	throttleKey := fmt.Sprintf("%s||%s", item.Endpoint, counter)

	// 3. 节流判断：10s内重复请求直接跳过，不执行SQL
	if !allowUpdateIndexDb(throttleKey) {
		return nil
	}
	// ===================== 新增：节流防抖逻辑 END =====================


	ts := item.Timestamp
	var endpointId int64 = -1
	// ===================== 1. 操作 endpoint 表（存储监控端点信息）=====================
	// 先查询端点是否已存在

	err := conn.QueryRow("SELECT id FROM endpoint WHERE endpoint = ?",
		item.Endpoint).Scan(&endpointId)

	// 端点不存在  查询异常：执行插入
	if err != nil || endpointId <= 0 {
		sqlStr := `INSERT INTO endpoint(endpoint, ts, t_create)	VALUES (?, ?, NOW())`
		result, err := conn.Exec(sqlStr, item.Endpoint, ts)
		if err != nil {
			log.Errorf("Insert endpoint err: %s", err)
			return err
		}
		endpointId, err = result.LastInsertId()
		if err != nil {
			log.Errorf("Get last insert id when insert endpoint err: %s", err)
			return err
		}
		if endpointId <= 0 {
			log.Errorf("insert to graph.endpoint, result LastInsertId is fail, endpoint=%s", item.Endpoint)
			return errors.New("insert to graph.endpoint failed")
		}
		// 上报endpoint表操作计数
		proc.IndexUpdateIncrDbEndpointInsertCnt.Incr()
	} else {

		// 端点已存在：更新最后上报时间、修改时间
		sqlStr := `UPDATE endpoint SET ts = ?, t_modify=NOW() WHERE id = ?`
		_, err := conn.Exec(sqlStr, ts, endpointId)
		if err != nil {
			log.Errorf("Update endpoint err: %s", err)
			return err
		}
		proc.IndexUpdateIncrDbEndpointUpdateCnt.Incr()
	}
	//end modified


	// 二次校验端点ID合法性
	if endpointId <= 0 {
		log.Errorf("no such endpoint in db, endpoint=%s", item.Endpoint)
		return errors.New("no such endpoint")
	}

	// tag_endpoint表
	// ===================== 2. 操作 tag_endpoint 表（存储端点+标签关联关系）=====================
	for tagKey, tagVal := range item.Tags {
		tag := fmt.Sprintf("%s=%s", tagKey, tagVal)
		var tag_endpoint_id int64 = -1

		// 查询标签+端点关联记录是否存在
		err := conn.QueryRow("SELECT id FROM tag_endpoint WHERE tag = ? and endpoint_id = ?",
			tag, endpointId).Scan(&tag_endpoint_id)

		// 记录不存在：插入新标签关联
		if err != nil || tag_endpoint_id <= 0 {
			sqlStr := `INSERT INTO tag_endpoint(tag, endpoint_id, ts, t_create)
				VALUES (?, ?, ?, NOW())`
			_, err := conn.Exec(sqlStr, tag, endpointId, ts)
			if err != nil {
				log.Errorf("Insert tag_endpoint err: %s", err)
				return err
			}

			// 上报tag_endpoint表操作计数
			proc.IndexUpdateIncrDbTagEndpointInsertCnt.Incr()

		} else {

			// 记录已存在：更新最后上报时间、修改时间
			sqlStr := `UPDATE tag_endpoint SET ts=?, t_modify=NOW() WHERE id = ?`
			_, err := conn.Exec(sqlStr, ts, tag_endpoint_id)
			if err != nil {
				log.Errorf("Update tag_endpoint err: %s", err)
				return err
			}
			proc.IndexUpdateIncrDbTagEndpointUpdateCnt.Incr()
		}
		//end modified
	}

	// endpoint_counter表
	// ===================== 3. 操作 endpoint_counter 表(存储完整指标标识：端点+指标+标签)=====================
	// 拼接完整counter名称：指标名 + 排序后的标签(标签有序保证唯一性)
	counter = item.Metric
	if len(item.Tags) > 0 {
		counter = fmt.Sprintf("%s/%s", counter, cutils.SortedTags(item.Tags))
	}

	// 查询该完整指标是否已存在
	var endpoint_counter_id int64 = -1
	err = conn.QueryRow("SELECT id FROM endpoint_counter WHERE counter = ? and endpoint_id = ?", counter, endpointId).Scan(&endpoint_counter_id)

	// 指标不存在：插入新记录
	if err != nil || endpoint_counter_id <= 0 {
		sqlStr := `INSERT INTO endpoint_counter(endpoint_id,counter,step,type,ts,t_create)
			VALUES (?,?,?,?,?,NOW())`
		_, err := conn.Exec(sqlStr, endpointId, counter, item.Step, item.DsType, ts)
		if err != nil {
			log.Errorf("Insert endpoint_counter err: %s", err)
			return err
		}

		if g.Config().Debug {
			log.Debugf("[DEBUG] insert into endpoint_counter endpointid: %d, COUNTER:" +
				"%s, ts: %d", endpointId, counter, ts)
		}

		// 上报endpoint_counter表操作计数
		proc.IndexUpdateIncrDbEndpointCounterInsertCnt.Incr()
	} else {

		// 指标已存在：更新上报时间、采集步长、数据类型
		sqlStr := `UPDATE endpoint_counter SET ts=?,step=?,type=?,t_modify=NOW() WHERE id = ?`
		_, err := conn.Exec(sqlStr, ts, item.Step, item.DsType, endpoint_counter_id)
		if err != nil {
			log.Errorf("Insert endpoint_counter err: %s", err)
			return err
		}
		if g.Config().Debug {
			log.Debugf("[DEBUG] upadte into endpoint_counter endpointid: %d, COUNTER:" +
				"%s, ts: %d", endpointId, counter, ts)
		}
		proc.IndexUpdateIncrDbEndpointCounterUpdateCnt.Incr()
	}
	//end modified

	return nil
}
