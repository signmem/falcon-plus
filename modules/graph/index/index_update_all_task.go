package index

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"time"

	nsema "github.com/toolkits/concurrent/semaphore"
	ntime "github.com/toolkits/time"

	cmodel "github.com/open-falcon/falcon-plus/common/model"
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/graph/g"
	proc "github.com/open-falcon/falcon-plus/modules/graph/proc"
)

const (
	DefaultUpdateStepInSec     = 2 * 24 * 3600 //更新步长,一定不能大于删除步长. 两天内的数据,都可以用来建立索引
	ConcurrentOfUpdateIndexAll = 1
)

var (
	semaIndexUpdateAllTask = nsema.NewSemaphore(ConcurrentOfUpdateIndexAll) //全量同步任务 并发控制器
	semaIndexUpdateAll     = nsema.NewSemaphore(4)                          // 索引全量更新时的mysql操作并发控制
)

// 索引全量更新的当前并行数
func GetConcurrentOfUpdateIndexAll() int {
	log.Debugf("[DEBUG] going to Flush GetConcurrentOfUpdateIndexAll()")
	return ConcurrentOfUpdateIndexAll - semaIndexUpdateAllTask.AvailablePermits()
}

// 索引的全量更新
func UpdateIndexAllByDefaultStep() {
	log.Debugf("[DEBUG] going to Flush UpdateIndexAllByDefaultStep()")
	UpdateIndexAll(DefaultUpdateStepInSec)
}

func UpdateIndexAll(updateStepInSec int64) {
	// 减少任务积压,但高并发时可能无效(AvailablePermits不是线程安全的)
	if semaIndexUpdateAllTask.AvailablePermits() <= 0 {
		log.Println("[DEBUG] updateIndexAll, concurrent not available")
		return
	}

	semaIndexUpdateAllTask.Acquire()
	defer semaIndexUpdateAllTask.Release()

	startTs := time.Now().Unix()
	cnt := updateIndexAll(updateStepInSec)
	endTs := time.Now().Unix()
	log.Printf("[DEBUG] UpdateIndexAll, lastStartTs %s, updateStepInSec %d, lastTimeConsumingInSec %d\n",
		ntime.FormatTs(startTs), updateStepInSec, endTs-startTs)

	// statistics
	proc.IndexUpdateAllCnt.SetCnt(int64(cnt))
	proc.IndexUpdateAll.Incr()
	proc.IndexUpdateAll.PutOther("lastStartTs", ntime.FormatTs(startTs))
	proc.IndexUpdateAll.PutOther("updateStepInSec", updateStepInSec)
	proc.IndexUpdateAll.PutOther("lastTimeConsumingInSec", endTs-startTs)
	proc.IndexUpdateAll.PutOther("updateCnt", cnt)
}

// 更新一条监控数据对应的索引. 用于手动添加索引,一般情况下不会使用
func UpdateIndexOne(endpoint string, metric string, tags map[string]string, dstype string, step int) error {
	itemDemo := &cmodel.GraphItem{
		Endpoint: endpoint,
		Metric:   metric,
		Tags:     tags,
		DsType:   dstype,
		Step:     step,
	}
	md5 := itemDemo.Checksum()
	uuid := itemDemo.UUID()

	cached := IndexedItemCache.Get(md5)
	if cached == nil {
		return fmt.Errorf("not found")
	}

	icitem := cached.(*IndexCacheItem)
	if icitem.UUID != uuid {
		return fmt.Errorf("bad type or step")
	}
	gitem := icitem.Item

	dbConn, err := g.GetDbConn("UpdateIndexIncrTask")
	if err != nil {
		log.Println("[ERROR] make dbConn fail", err)
		return err
	}

	return updateIndexFromOneItem(gitem, dbConn)
}

func updateIndexAll(updateStepInSec int64) int {
	var ret int = 0
	if IndexedItemCache == nil || IndexedItemCache.Size() <= 0 {
		return ret
	}

	dbConn, err := g.GetDbConn("UpdateIndexIncrTask")
	if err != nil {
		log.Println("[ERROR] make dbConn fail", err)
		return ret
	}

	// lastTs for update index
	ts := time.Now().Unix()
	lastTs := ts - updateStepInSec

	keys := IndexedItemCache.Keys()
	for _, key := range keys {
		icitem := IndexedItemCache.Get(key)
		if icitem == nil {
			continue
		}

		if g.Config().Debug {
			log.Printf("[DEBUG] key from IndexedItemCache: %s", key)
		}

		gitem := icitem.(*IndexCacheItem).Item
		if gitem.Timestamp < lastTs { //缓存中的数据太旧了,不能用于索引的全量更新

			if g.Config().Debug {
				log.Debugf("[DEBUG] remove from IndexedItemCache: metric %s, endpoint %s, time %d ",
					gitem.Metric, gitem.Endpoint, gitem.Timestamp)
			}
			IndexedItemCache.Remove(key) //在这里做个删除,有点恶心
			continue
		}
		// 并发写mysql
		semaIndexUpdateAll.Acquire()
		go func(gitem *cmodel.GraphItem, dbConn *sql.DB) {
			defer semaIndexUpdateAll.Release()
			err := updateIndexFromOneItem(gitem, dbConn)
			if g.Config().Debug {
				log.Debugf("[DEBUG] update from semaIndexUpdateAll: metric %s, endpoint %s, time %d ",
					gitem.Metric, gitem.Endpoint, gitem.Timestamp)
			}
			if err != nil {
				proc.IndexUpdateAllErrorCnt.Incr()
			}
		}(gitem, dbConn)

		ret++
	}

	return ret
}

// 根据item,更新mysql
func updateIndexFromOneItem(item *cmodel.GraphItem, conn *sql.DB) error {
	if item == nil {
		return nil
	}

	ts := item.Timestamp
	var endpointId int64 = -1

	// endpoint表
	//vincent.zhang修改，解决updateAll时graph db中各个表id增长过快问题
	/*
		sqlStr := `INSERT INTO endpoint(endpoint, ts, t_create)
			VALUES (?, ?, NOW())
			ON DUPLICATE KEY UPDATE ts=?, t_modify=NOW()`

		_, err := conn.Exec(sqlStr, item.Endpoint, ts, ts)
		if err != nil {
			log.Error(err)
			return err
		}*/
	err := conn.QueryRow("SELECT id FROM endpoint WHERE endpoint = ?",
		item.Endpoint).Scan(&endpointId)
	/*
		if err != nil {
			log.Infof("select id from endpoint error, endpoint: %s, err: %s", item.Endpoint, err)
		}
	*/
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
	} else {
		sqlStr := `UPDATE endpoint SET ts = ?, t_modify=NOW() WHERE id = ?`
		_, err := conn.Exec(sqlStr, ts, endpointId)
		if err != nil {
			log.Errorf("Update endpoint err: %s", err)
			return err
		}
	}
	//end modified

	proc.IndexUpdateIncrDbEndpointInsertCnt.Incr()
	//vincent.zhang修改，解决updateAll时graph db中各个表id增长过快问题
	/*
		err = conn.QueryRow("SELECT id FROM endpoint WHERE endpoint = ?", item.Endpoint).Scan(&endpointId)
		if err != nil {
			log.Error(err)
			return err
		}*/
	//end modified
	if endpointId <= 0 {
		log.Errorf("no such endpoint in db, endpoint=%s", item.Endpoint)
		return errors.New("no such endpoint")
	}

	// tag_endpoint表
	for tagKey, tagVal := range item.Tags {
		tag := fmt.Sprintf("%s=%s", tagKey, tagVal)
		//vincent.zhang修改，解决updateAll时graph db中各个表id增长过快问题
		/*
			sqlStr := `INSERT INTO tag_endpoint(tag, endpoint_id, ts, t_create)
				VALUES (?, ?, ?, NOW())
				ON DUPLICATE KEY UPDATE ts=?, t_modify=NOW()`

			_, err := conn.Exec(sqlStr, tag, endpointId, ts, ts)
			if err != nil {
				log.Error(err)
				return err
			}*/
		var tag_endpoint_id int64 = -1
		err := conn.QueryRow("SELECT id FROM tag_endpoint WHERE tag = ? and endpoint_id = ?",
			tag, endpointId).Scan(&tag_endpoint_id)
		/*
			if err != nil {
				log.Infof("select id from tag_endpoint error, endpoint_id: %d, tag: %s,  err: %s", endpointId, tag, err)
			}
		*/
		if err != nil || tag_endpoint_id <= 0 {
			sqlStr := `INSERT INTO tag_endpoint(tag, endpoint_id, ts, t_create)
				VALUES (?, ?, ?, NOW())`
			_, err := conn.Exec(sqlStr, tag, endpointId, ts)
			if err != nil {
				log.Errorf("Insert tag_endpoint err: %s", err)
				return err
			}


		} else {
			sqlStr := `UPDATE tag_endpoint SET ts=?, t_modify=NOW() WHERE id = ?`
			_, err := conn.Exec(sqlStr, ts, tag_endpoint_id)
			if err != nil {
				log.Errorf("Update tag_endpoint err: %s", err)
				return err
			}
		}
		//end modified

		proc.IndexUpdateIncrDbTagEndpointInsertCnt.Incr()
	}

	// endpoint_counter表
	counter := item.Metric
	if len(item.Tags) > 0 {
		counter = fmt.Sprintf("%s/%s", counter, cutils.SortedTags(item.Tags))
	}
	//vincent.zhang修改，解决updateAll时graph db中各个表id增长过快问题
	/*
		sqlStr = `INSERT INTO endpoint_counter(endpoint_id,counter,step,type,ts,t_create)
			VALUES (?,?,?,?,?,NOW())
			ON DUPLICATE KEY UPDATE ts=?,step=?,type=?,t_modify=NOW()`

		_, err = conn.Exec(sqlStr, endpointId, counter, item.Step, item.DsType, ts, ts, item.Step, item.DsType)
		if err != nil {
			log.Error(err)
			return err
		}
	*/
	var endpoint_counter_id int64 = -1
	err = conn.QueryRow("SELECT id FROM endpoint_counter WHERE counter = ? and endpoint_id = ?", counter, endpointId).Scan(&endpoint_counter_id)
	/*
		if err != nil {
			log.Infof("select id from endpoint_counter error, endpoint_id: %d, counter: %s, err: %s", endpointId, counter, err)
		}
	*/
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

	} else {
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
	}
	//end modified

	proc.IndexUpdateIncrDbEndpointCounterInsertCnt.Incr()

	return nil
}
