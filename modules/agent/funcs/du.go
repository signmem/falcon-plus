package funcs

import (
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/toolkits/sys"
	"log"
	"strconv"
	"strings"
)

func DuMetrics() (L []*model.MetricValue) {
	paths := g.DuPaths()
	for _, path := range paths {
		out, err := sys.CmdOutNoLn("du", "-bs", path)
		if err != nil {
			log.Println("du -bs", path, "fail", err)
			continue
		}

		arr := strings.Fields(out)
		if len(arr) == 1 {
			continue
		}

		size, err := strconv.ParseUint(arr[0], 10, 64)
		if err != nil {
			log.Println("cannot parse du -bs", path, "output")
			continue
		}

		L = append(L, GaugeValue(g.DU_BS, size, "path="+path))
	}

	return
}
