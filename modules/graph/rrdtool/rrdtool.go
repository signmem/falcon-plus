package rrdtool

import (
	"errors"
	"log"
	"math"
	"sync/atomic"
	"time"

	cmodel "github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/rrdlite"
	"github.com/toolkits/file"

	"github.com/open-falcon/falcon-plus/modules/graph/g"
	"github.com/open-falcon/falcon-plus/modules/graph/store"
)

var (
	disk_counter uint64
	net_counter  uint64
)

type fetch_t struct {
	filename string
	cf       string
	start    int64
	end      int64
	step     int
	data     []*cmodel.RRDData
}

type flushfile_t struct {
	filename string
	items    []*cmodel.GraphItem
}

type readfile_t struct {
	filename string
	data     []byte
}

func Start() {
	cfg := g.Config()
	var err error
	// check data dir
	if err = file.EnsureDirRW(cfg.RRD.Storage); err != nil {
		log.Fatalln("rrdtool.Start error, bad data dir "+cfg.RRD.Storage+",", err)
	}

	migrate_start(cfg)

	// sync disk
	go syncDisk()
	go ioWorker()
	log.Println("rrdtool.Start ok")
}

// RRA.Point.Size
const (
	RRA5sPointCnt  = 259200 //5s原始一个点存15d
	RRA10sPointCnt = 129600 //10s原始一个点存15d
	RRA30sPointCnt = 43200  //30s原始一个点存15d
	RRA1mPointCnt  = 21600  //1m原始一个点存15d
	RRA5mPointCnt  = 4320   //5m原始一个点存15d
	RRA1hPointCnt  = 360    //1h原始一个点存15d

	RRA20PointCnt  = 4320 // 20m一个点存60d
	RRA180PointCnt = 720  // 3h一个点存90d
	RRA720PointCnt = 730  // 12h一个点存1year
)

func create(filename string, item *cmodel.GraphItem) error {
	now := time.Now()
	start := now.Add(time.Duration(-24) * time.Hour)
	step := uint(item.Step)
	//log.Printf("create rrd filename: %s, step is %u \n", filename, step)
	c := rrdlite.NewCreator(filename, start, step)
	c.DS("metric", item.DsType, item.Heartbeat, item.Min, item.Max)
	//log.Printf("create rrd: metric, %s, %d, %s, %s \n", item.DsType, item.Heartbeat, item.Min, item.Max)
	//add by hanyun
	if step == 5 {
		//log.Println("start create rrd rra, step is 5")
		c.RRA("AVERAGE", 0, 1, RRA5sPointCnt)

		c.RRA("AVERAGE", 0, 240, RRA20PointCnt)
		c.RRA("MAX", 0, 240, RRA20PointCnt)
		c.RRA("MIN", 0, 240, RRA20PointCnt)

		c.RRA("AVERAGE", 0, 2160, RRA180PointCnt)
		c.RRA("MAX", 0, 2160, RRA180PointCnt)
		c.RRA("MIN", 0, 2160, RRA180PointCnt)

		c.RRA("AVERAGE", 0, 8640, RRA720PointCnt)
		c.RRA("MAX", 0, 8640, RRA720PointCnt)
		c.RRA("MIN", 0, 8640, RRA720PointCnt)
		//log.Println("finish create rrd rra, step is 5")
	} else if step == 10 {
		//log.Println("start create rrd rra, step is 10")
		c.RRA("AVERAGE", 0, 1, RRA10sPointCnt)

		c.RRA("AVERAGE", 0, 120, RRA20PointCnt)
		c.RRA("MAX", 0, 120, RRA20PointCnt)
		c.RRA("MIN", 0, 120, RRA20PointCnt)

		c.RRA("AVERAGE", 0, 1080, RRA180PointCnt)
		c.RRA("MAX", 0, 1080, RRA180PointCnt)
		c.RRA("MIN", 0, 1080, RRA180PointCnt)

		c.RRA("AVERAGE", 0, 4320, RRA720PointCnt)
		c.RRA("MAX", 0, 4320, RRA720PointCnt)
		c.RRA("MIN", 0, 4320, RRA720PointCnt)
		//log.Println("finish create rrd rra, step is 10")
	} else if step == 30 {
		//log.Println("start create rrd rra, step is 30")
		c.RRA("AVERAGE", 0, 1, RRA30sPointCnt)

		c.RRA("AVERAGE", 0, 40, RRA20PointCnt)
		c.RRA("MAX", 0, 40, RRA20PointCnt)
		c.RRA("MIN", 0, 40, RRA20PointCnt)

		c.RRA("AVERAGE", 0, 360, RRA180PointCnt)
		c.RRA("MAX", 0, 360, RRA180PointCnt)
		c.RRA("MIN", 0, 360, RRA180PointCnt)

		c.RRA("AVERAGE", 0, 1440, RRA720PointCnt)
		c.RRA("MAX", 0, 1440, RRA720PointCnt)
		c.RRA("MIN", 0, 1440, RRA720PointCnt)
		//log.Println("finish create rrd rra, step is 30")
	} else if step == 60 {
		//log.Println("start create rrd rra, step is 60")
		c.RRA("AVERAGE", 0, 1, RRA1mPointCnt)

		c.RRA("AVERAGE", 0, 20, RRA20PointCnt)
		c.RRA("MAX", 0, 20, RRA20PointCnt)
		c.RRA("MIN", 0, 20, RRA20PointCnt)

		c.RRA("AVERAGE", 0, 180, RRA180PointCnt)
		c.RRA("MAX", 0, 180, RRA180PointCnt)
		c.RRA("MIN", 0, 180, RRA180PointCnt)

		c.RRA("AVERAGE", 0, 720, RRA720PointCnt)
		c.RRA("MAX", 0, 720, RRA720PointCnt)
		c.RRA("MIN", 0, 720, RRA720PointCnt)
		//log.Println("finish create rrd rra, step is 60")
	} else if step == 300 {
		//log.Println("start create rrd rra, step is 300")
		c.RRA("AVERAGE", 0, 1, RRA5mPointCnt)

		c.RRA("AVERAGE", 0, 4, RRA20PointCnt)
		c.RRA("MAX", 0, 4, RRA20PointCnt)
		c.RRA("MIN", 0, 4, RRA20PointCnt)

		c.RRA("AVERAGE", 0, 36, RRA180PointCnt)
		c.RRA("MAX", 0, 36, RRA180PointCnt)
		c.RRA("MIN", 0, 36, RRA180PointCnt)

		c.RRA("AVERAGE", 0, 144, RRA720PointCnt)
		c.RRA("MAX", 0, 144, RRA720PointCnt)
		c.RRA("MIN", 0, 144, RRA720PointCnt)
		//log.Println("finish create rrd rra, step is 300")
	} else if step == 3600 {
		//log.Println("start create rrd rra, step is 3600")
		c.RRA("AVERAGE", 0, 1, RRA1hPointCnt)

		c.RRA("AVERAGE", 0, 3, RRA180PointCnt)
		c.RRA("MAX", 0, 3, RRA180PointCnt)
		c.RRA("MIN", 0, 3, RRA180PointCnt)

		c.RRA("AVERAGE", 0, 12, RRA720PointCnt)
		c.RRA("MAX", 0, 12, RRA720PointCnt)
		c.RRA("MIN", 0, 12, RRA720PointCnt)
		//log.Println("finish create rrd rra, step is 3600")
	} else {
		log.Printf("create rrd rra, default step 60s \n")
		c.RRA("AVERAGE", 0, 1, RRA1mPointCnt)

		c.RRA("AVERAGE", 0, 20, RRA20PointCnt)
		c.RRA("MAX", 0, 20, RRA20PointCnt)
		c.RRA("MIN", 0, 20, RRA20PointCnt)

		c.RRA("AVERAGE", 0, 180, RRA180PointCnt)
		c.RRA("MAX", 0, 180, RRA180PointCnt)
		c.RRA("MIN", 0, 180, RRA180PointCnt)

		c.RRA("AVERAGE", 0, 720, RRA720PointCnt)
		c.RRA("MAX", 0, 720, RRA720PointCnt)
		c.RRA("MIN", 0, 720, RRA720PointCnt)

	}

	/*
		// 设置各种归档策略
		// 1分钟一个点存 12小时
		c.RRA("AVERAGE", 0, 1, RRA1PointCnt)

		// 5m一个点存2d
		c.RRA("AVERAGE", 0, 5, RRA5PointCnt)
		c.RRA("MAX", 0, 5, RRA5PointCnt)
		c.RRA("MIN", 0, 5, RRA5PointCnt)

		// 20m一个点存7d
		c.RRA("AVERAGE", 0, 20, RRA20PointCnt)
		c.RRA("MAX", 0, 20, RRA20PointCnt)
		c.RRA("MIN", 0, 20, RRA20PointCnt)

		// 3小时一个点存3个月
		c.RRA("AVERAGE", 0, 180, RRA180PointCnt)
		c.RRA("MAX", 0, 180, RRA180PointCnt)
		c.RRA("MIN", 0, 180, RRA180PointCnt)

		// 12小时一个点存1year
		c.RRA("AVERAGE", 0, 720, RRA720PointCnt)
		c.RRA("MAX", 0, 720, RRA720PointCnt)
		c.RRA("MIN", 0, 720, RRA720PointCnt)
	*/
	return c.Create(true)
}

func update(filename string, items []*cmodel.GraphItem) error {
	u := rrdlite.NewUpdater(filename)

	for _, item := range items {
		v := math.Abs(item.Value)
		if v > 1e+300 || (v < 1e-300 && v > 0) {
			continue
		}
		if item.DsType == "DERIVE" || item.DsType == "COUNTER" {
			u.Cache(item.Timestamp, int(item.Value))
		} else {
			u.Cache(item.Timestamp, item.Value)
		}
	}

	return u.Update()
}

// flush to disk from memory
// 最新的数据在列表的最后面
// TODO fix me, filename fmt from item[0], it's hard to keep consistent
func flushrrd(filename string, items []*cmodel.GraphItem) error {
	if items == nil || len(items) == 0 {
		return errors.New("empty items")
	}

	if !g.IsRrdFileExist(filename) {
		baseDir := file.Dir(filename)

		err := file.InsureDir(baseDir)
		if err != nil {
			return err
		}

		err = create(filename, items[0])
		if err != nil {
			return err
		}
	}

	return update(filename, items)
}

func ReadFile(filename string) ([]byte, error) {
	done := make(chan error, 1)
	task := &io_task_t{
		method: IO_TASK_M_READ,
		args:   &readfile_t{filename: filename},
		done:   done,
	}

	io_task_chan <- task
	err := <-done
	return task.args.(*readfile_t).data, err
}

func FlushFile(filename string, items []*cmodel.GraphItem) error {
	done := make(chan error, 1)
	io_task_chan <- &io_task_t{
		method: IO_TASK_M_FLUSH,
		args: &flushfile_t{
			filename: filename,
			items:    items,
		},
		done: done,
	}
	atomic.AddUint64(&disk_counter, 1)
	return <-done
}

func Fetch(filename string, cf string, start, end int64, step int) ([]*cmodel.RRDData, error) {
	done := make(chan error, 1)
	task := &io_task_t{
		method: IO_TASK_M_FETCH,
		args: &fetch_t{
			filename: filename,
			cf:       cf,
			start:    start,
			end:      end,
			step:     step,
		},
		done: done,
	}
	io_task_chan <- task
	err := <-done
	return task.args.(*fetch_t).data, err
}

func fetch(filename string, cf string, start, end int64, step int) ([]*cmodel.RRDData, error) {
	start_t := time.Unix(start, 0)
	end_t := time.Unix(end, 0)
	step_t := time.Duration(step) * time.Second

	fetchRes, err := rrdlite.Fetch(filename, cf, start_t, end_t, step_t)
	if err != nil {
		return []*cmodel.RRDData{}, err
	}

	defer fetchRes.FreeValues()

	values := fetchRes.Values()
	size := len(values)
	ret := make([]*cmodel.RRDData, size)

	start_ts := fetchRes.Start.Unix()
	step_s := fetchRes.Step.Seconds()

	for i, val := range values {
		ts := start_ts + int64(i+1)*int64(step_s)
		d := &cmodel.RRDData{
			Timestamp: ts,
			Value:     cmodel.JsonFloat(val),
		}
		ret[i] = d
	}

	return ret, nil
}

func FlushAll(force bool) {
	n := store.GraphItems.Size / 10
	for i := 0; i < store.GraphItems.Size; i++ {
		FlushRRD(i, force)
		if i%n == 0 {
			log.Printf("flush hash idx:%03d size:%03d disk:%08d net:%08d\n",
				i, store.GraphItems.Size, disk_counter, net_counter)
		}
	}
	log.Printf("flush hash done (disk:%08d net:%08d)\n", disk_counter, net_counter)
}

func CommitByKey(key string) {

	md5, dsType, step, err := g.SplitRrdCacheKey(key)
	if err != nil {
		return
	}
	filename := g.RrdFileName(g.Config().RRD.Storage, md5, dsType, step)

	items := store.GraphItems.PopAll(key)
	if len(items) == 0 {
		return
	}
	FlushFile(filename, items)
}

func PullByKey(key string) {
	done := make(chan error)

	item := store.GraphItems.First(key)
	if item == nil {
		return
	}
	node, err := Consistent.Get(item.PrimaryKey())
	if err != nil {
		return
	}
	Net_task_ch[node] <- &Net_task_t{
		Method: NET_TASK_M_PULL,
		Key:    key,
		Done:   done,
	}
	// net_task slow, shouldn't block syncDisk() or FlushAll()
	// warning: recev sigout when migrating, maybe lost memory data
	go func() {
		err := <-done
		if err != nil {
			log.Printf("get %s from remote err[%s]\n", key, err)
			return
		}
		atomic.AddUint64(&net_counter, 1)
		//todo: flushfile after getfile? not yet
	}()
}

func FlushRRD(idx int, force bool) {
	begin := time.Now()
	atomic.StoreInt32(&flushrrd_timeout, 0)

	keys := store.GraphItems.KeysByIndex(idx)
	if len(keys) == 0 {
		return
	}

	for _, key := range keys {
		flag, _ := store.GraphItems.GetFlag(key)

		//write err data to local filename
		if force == false && g.Config().Migrate.Enabled && flag&g.GRAPH_F_MISS != 0 {
			if time.Since(begin) > time.Millisecond*g.FLUSH_DISK_STEP {
				atomic.StoreInt32(&flushrrd_timeout, 1)
			}
			PullByKey(key)
		} else {
			CommitByKey(key)
		}
	}
}
