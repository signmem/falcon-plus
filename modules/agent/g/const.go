package g

import (
	"time"
)

// changelog:
// 3.1.3: code refactor
// 3.1.4: bugfix ignore configuration
// 5.0.0: 支持通过配置控制是否开启/run接口；收集udp流量数据；du某个目录的大小
// 5.1.0: 同步插件的时候不再使用checksum机制
// 5.1.1: 修复往多个transfer发送数据的时候crash的问题
// 5.1.2: ignore mount point when blocks=0
// 5.1.6: update git version update method to get filecheck version, use rsync to update plugin
// 5.2.0: update plugin, update rpm automatic, disable git function
// 5.2.1
//      fix get ip addr function,
//      support nvme device,
//      support root startup
//      rsync with password
//      add  local:22230/v1/ipadd  api
//      update localip every day
//
// 5.2.2
//     fix plugin crush without bash header
//     disable set GID error msg
//     fix centos5 portmetric() error
//     version file location -> version/version
//     disable   localhost:22230/v1/run  api
//     rewrite log function
//     add tags="pool=deploypoolname"
//
//
// 5.3.1 
//    支持 kubevirt 增加  tags -> family=virtul 功能   
//    对 22230/v1/push 接口上报数据进行 tags, metric 字符标准化
//    增加检测 /etc/vip-vm-monitor.conf 是否对 swap 进行监控


const (
	VERSION          = "5.3.1_20221130"
	COLLECT_INTERVAL = time.Second
	URL_CHECK_HEALTH = "url.check.health"
	NET_PORT_LISTEN  = "net.port.listen"
	DU_BS            = "du.bs"
	PROC_NUM         = "proc.num"
)
