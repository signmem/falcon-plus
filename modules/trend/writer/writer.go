package writer

import (
	"time"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
	"github.com/open-falcon/falcon-plus/modules/trend/proc"
)

const (
	DEFAULT_DB_WRITE_BATCH          = 200
	DEFAULT_DB_WRITE_BATCH_INTERVAL = 200 //默认200ms
)

var (
	DBWriteBatch         int
	DBWriteBatchInterval int
)

func Start() {
	initDB()
	cfg := g.Config()
	DBWriteBatch = cfg.DB.Batch

	DBWriteBatchInterval = cfg.DB.BatchInterval
	if DBWriteBatch <= 0 {
		DBWriteBatch = DEFAULT_DB_WRITE_BATCH
	}

	if DBWriteBatchInterval <= 0 {
		DBWriteBatchInterval = DEFAULT_DB_WRITE_BATCH_INTERVAL
	}
}

func Close() {
	closeDB()
}

func FlushToDB(key int64, items []*g.TrendResult) {

	if g.Config().DBLog {
		g.Logger.Debug("[DBLOG] ===> FlushToDB")
	}

	ts := key * g.TREND_INTERVALS
	length := len(items)
	if length == 0 {
		g.Logger.Debug("FlushToDB() store data length is zero")
		return
	}
	a := length / DBWriteBatch
	b := length % DBWriteBatch
	interval := time.Duration(DBWriteBatchInterval) * time.Millisecond
	for i := 0; i < a; i++ {
		storeTrend(ts, items[DBWriteBatch*i:DBWriteBatch*(i+1)])
		time.Sleep(interval)
	}
	if b > 0 {
		storeTrend(ts, items[DBWriteBatch*a:])
		time.Sleep(interval)
	}
}

func storeTrend(ts int64, items []*g.TrendResult) {
	if g.Config().Debug {
		g.Logger.Debug("storeTrend ===>")
		g.Logger.Debug(items)
	}
	length := len(items)
	if len(items) == 0 {
		g.Logger.Warning("storeTrend() store data length is zero")
		return
	}
	if DB == nil {
		g.Logger.Error("storeTrend() DB instance is nil")
		proc.WriteToDBFailCnt.IncrBy(int64(length))
		return
	}
	tx, _ := DB.Begin()

	//metric_stmt, _ := tx.Prepare("insert into host_metrics(hostname, metric, tags, dstype, step, create_time) values (?, ?, ?, ?, ?, NOW()) on duplicate key update dstype=?, step=?")
	metric_stmt, _ := tx.Prepare("insert into host_metrics(hostname, metric, tags, dstype, step, create_time) values (?, ?, ?, ?, ?, NOW())")
	update_stmt, _ := tx.Prepare("update host_metrics set dstype=?, step=? where id=?")
	defer update_stmt.Close()
	query_stmt, _ := tx.Prepare("select id, dstype, step from host_metrics where hostname = ? and metric = ? and tags = ?")
	defer query_stmt.Close()
	trend_stmt, _ := tx.Prepare("insert into trend(metric_id, ts, min, avg, max, num, create_time) values (?, ?, ?, ?, ?, ?, NOW())")
	defer trend_stmt.Close()
	for _, item := range items {
		if item == nil {
			proc.WriteToDBFailCnt.Incr()
			continue
		}
		var metric_id int64
		var metric_dstype string
		var metric_step int
		metric_err := query_stmt.QueryRow(item.Endpoint, item.Metric, item.Tags).Scan(&metric_id, &metric_dstype, &metric_step)
		if metric_err != nil || metric_id <= 0 {
			if metric_err != nil {
				g.Logger.Errorf("select id, dstype, step from host_metrics table error: %s", metric_err.Error())
			}
			g.Logger.Debugf("no such metric in db, prepare for inserting, metric=%s/%s/%s", item.Endpoint, item.Metric, item.Tags)
			result, err := metric_stmt.Exec(item.Endpoint, item.Metric, item.Tags, item.DsType, item.Step)
			if err != nil {
				g.Logger.Errorf("insert metric=%s/%s/%s to host_metrics table error: %s", item.Endpoint, item.Metric, item.Tags, err.Error())
				proc.WriteToDBFailCnt.Incr()
				continue
			}
			metric_id, metric_err = result.LastInsertId()
			if metric_err != nil {
				g.Logger.Errorf("insert metric=%s/%s/%s result LastInsertId error: %s", item.Endpoint, item.Metric, item.Tags, metric_err.Error())
				proc.WriteToDBFailCnt.Incr()
				continue
			}
			if metric_id <= 0 {
				g.Logger.Errorf("insert metric=%s/%s/%s result LastInsertId is fail, metric_id=%d", item.Endpoint, item.Metric, item.Tags, metric_id)
				proc.WriteToDBFailCnt.Incr()
				continue
			}
		} else {
			if metric_dstype != item.DsType || metric_step != item.Step {
				g.Logger.Debugf("update host_metrics table, metric_id: %d, %s=>%s, %d=>%d", metric_id, metric_dstype, item.DsType, metric_step, item.Step)
				_, err := update_stmt.Exec(item.DsType, item.Step, metric_id)
				if err != nil {
					g.Logger.Errorf("update host_metrics table, metric_id: %d, %s=>%s, %d=>%d error: %s", metric_id, metric_dstype, item.DsType, metric_step, item.Step, err.Error())
				}
			}
		}

		_, err := trend_stmt.Exec(metric_id, ts, item.Min, item.Avg, item.Max, item.Num)
		if err != nil {
			g.Logger.Errorf("insert trend error: %s", err.Error())
			proc.WriteToDBFailCnt.Incr()
		}
		proc.WriteToDBCnt.Incr()
	}

	err := tx.Commit()
	if err != nil {
		g.Logger.Errorf("storeTrend() commit is fail: %s", err.Error())
	}
}
