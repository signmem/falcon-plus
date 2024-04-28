package cron

import (
	"github.com/open-falcon/falcon-plus/modules/graph/g"
	"github.com/open-falcon/falcon-plus/modules/graph/index"
	"time"
)

func Flushcache() {
	ticker := time.NewTicker(time.Duration(g.Config().FlustInterval) * time.Second)
	for {
		<-ticker.C
		go index.UpdateIndexAllByDefaultStep()
		go index.GetConcurrentOfUpdateIndexAll()
	}
}

