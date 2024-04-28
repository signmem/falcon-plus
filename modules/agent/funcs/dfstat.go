package funcs

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"github.com/toolkits/nux"
	"regexp"
)

func DeviceMetrics() (L []*model.MetricValue) {
	mountPoints, err := nux.ListMountPoint()

	if err != nil {
		log.Error("collect device metrics fail:", err)
		return
	}

	var myMountPoints map[string]bool = make(map[string]bool)

	if len(g.Config().Collector.MountPoint) > 0 {
		for _, mp := range g.Config().Collector.MountPoint {
			myMountPoints[mp] = true
		}
	}

	var diskTotal uint64 = 0
	var diskUsed uint64 = 0

	for idx := range mountPoints {
		fsSpec, fsFile, fsVfstype := mountPoints[idx][0], mountPoints[idx][1], mountPoints[idx][2]
		if len(myMountPoints) > 0 {
			if _, ok := myMountPoints[fsFile]; !ok {
				log.Debug("mount point not matched with config", fsFile, "ignored.")
				continue
			}
		}

		var du *nux.DeviceUsage
		du, err = nux.BuildDeviceUsage(fsSpec, fsFile, fsVfstype)
		if err != nil {
			log.Error("DeviceMetrics() error ", err)
			continue
		}

		if du.BlocksAll == 0 {
			continue
		}

		diskTotal += du.BlocksAll
		diskUsed += du.BlocksUsed

		/*
		2021-08-11  edit by terry.zeng
		filter fstype allow "xfs|^ext3|^ext4|^ntfs|^vfat|^fat32"
		filter mountpoint  deny "^/run|^/sys|^/proc|^tmpfs|^cgroup|^devtmpfs|^mqueue"
		*/

		fsString := "xfs|^ext3|^ext4|^ntfs|^vfat|^fat32"
		fsStatus, _ := regexp.MatchString(fsString, du.FsVfstype)
		if fsStatus == false {
			continue
		}

		mountString := "^/run|^/sys|^/proc|^tmpfs|^cgroup|^devtmpfs|^mqueue"
		mountStatus, _ := regexp.MatchString(mountString, du.FsFile)
		if mountStatus == true {
			continue
		}

		tags := fmt.Sprintf("mount=%s,fstype=%s", du.FsFile, du.FsVfstype)
		L = append(L, GaugeValue("df.bytes.total", du.BlocksAll, tags))
		L = append(L, GaugeValue("df.bytes.used", du.BlocksUsed, tags))
		L = append(L, GaugeValue("df.bytes.free", du.BlocksFree, tags))
		L = append(L, GaugeValue("df.bytes.used.percent", du.BlocksUsedPercent, tags))
		L = append(L, GaugeValue("df.bytes.free.percent", du.BlocksFreePercent, tags))
		L = append(L, GaugeValue("df.inodes.total", du.InodesAll, tags))
		L = append(L, GaugeValue("df.inodes.used", du.InodesUsed, tags))
		L = append(L, GaugeValue("df.inodes.free", du.InodesFree, tags))
		L = append(L, GaugeValue("df.inodes.used.percent", du.InodesUsedPercent, tags))
		L = append(L, GaugeValue("df.inodes.free.percent", du.InodesFreePercent, tags))

	}

	if len(L) > 0 && diskTotal > 0 {
		L = append(L, GaugeValue("df.statistics.total", float64(diskTotal)))
		L = append(L, GaugeValue("df.statistics.used", float64(diskUsed)))
		L = append(L, GaugeValue("df.statistics.used.percent", float64(diskUsed)*100.0/float64(diskTotal)))
	}

	return
}
