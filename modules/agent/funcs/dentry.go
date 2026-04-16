package funcs

import (
	"github.com/signmem/falcon-plus/common/model"
	"log"
	"github.com/signmem/nux"
)

func DentryMetrics() []*model.MetricValue {

	m, err := nux.DentryInfo()
	if err != nil {
		log.Println("DentryInfo() error", err)
		return nil
	}

	return []*model.MetricValue{
		GaugeValue("dentry.nr_dentry", m.NR_DENTRY),
		GaugeValue("dentry.nr_unused", m.NR_UNUSED),
		GaugeValue("dentry.age_limit", m.AGE_LIMIT),
		GaugeValue("dentry.want_pages", m.WANT_PAGES),
		GaugeValue("dentry.nr_negative", m.NR_NEGATIVE),
		GaugeValue("dentry.dummy", m.DUMMY),
	}
}