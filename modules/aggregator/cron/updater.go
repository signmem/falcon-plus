package cron

import (
	"github.com/signmem/falcon-plus/modules/aggregator/db"
	"github.com/signmem/falcon-plus/modules/aggregator/g"
	"time"
)

func UpdateItems() {
	for {
		updateItems()
		d := time.Duration(g.Config().Database.Interval) * time.Second
		time.Sleep(d)
	}
}

func updateItems() {
	items, err := db.ReadClusterMonitorItems()
	if err != nil {
		return
	}

	deleteNoUseWorker(items)
	createWorkerIfNeed(items)
}
