go 1.18

module github.com/signmem/falcon-plus

replace github.com/ugorji/go v0.0.0-20171122102828-84cb69a8af83 => github.com/ugorji/go/codec v1.2.12

replace github.com/ugorji/go/codec v1.2.12 => github.com/ugorji/go v1.2.12

replace github.com/Shopify/sarama => github.com/IBM/sarama v1.43.3

replace github.com/IBM/sarama => github.com/Shopify/sarama v1.43.3

replace github.com/gomodule/redigo/redis v0.0.1 => github.com/gomodule/redigo v2.0.0+incompatible

require (
	github.com/DeanThompson/ginpprof v0.0.0-20201112072838-007b1e56b2e1
	github.com/ShellCode33/VM-Detection v0.1.0
	github.com/Shopify/sarama v0.0.0-00010101000000-000000000000
	github.com/astaxie/beego v1.10.1
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/coreos/go-log v0.0.0-20180308165134-b22fd89e1882
	github.com/emirpasic/gods v1.18.1
	github.com/facebookgo/atomicfile v0.0.0-20151019160806-2de1f203e7d5
	github.com/garyburd/redigo v1.6.4
	github.com/gin-gonic/gin v1.10.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/gomodule/redigo v1.9.2
	github.com/jinzhu/gorm v1.9.16
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/masato25/yaag v0.0.0-20170704095552-00862ec4db8e
	github.com/niean/goperfcounter v0.0.0-20160108100052-24860a8d3fac
	github.com/redis/go-redis/v9 v9.7.0
	github.com/satori/go.uuid v1.2.0
	github.com/signmem/consistent v0.0.0-20240428121820-de78f9192db8
	github.com/signmem/file v0.0.0-20241205062615-8a57adbcc6d5
	github.com/signmem/go-log v0.0.0-20240403024112-86a3635680b9
	github.com/signmem/redislock v0.0.0-20240429073207-03e83a3aa6eb
	github.com/signmem/sarama v0.0.0-20241204033045-f7fcfef3ca58
	github.com/sirupsen/logrus v1.9.3
	github.com/smartystreets/goconvey v1.8.1
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/tatsushid/go-fastping v0.0.0-20160109021039-d7bb493dee3e
	github.com/toolkits/cache v0.0.0-20190218093630-cfb07b7585e5
	github.com/toolkits/concurrent v0.0.0-20150624120057-a4371d70e3e3
	github.com/toolkits/conn_pool v0.0.0-20170512061817-2b758bec1177
	github.com/toolkits/consistent v0.0.0-20150827090850-a6f56a64d1b1
	github.com/toolkits/container v0.0.0-20151219225805-ba7d73adeaca
	github.com/toolkits/core v0.0.0-20141116054942-0ebf14900fe2
	github.com/toolkits/cron v0.0.0-20150624115642-bebc2953afa6
	github.com/toolkits/file v0.0.0-20160325033739-a5b3c5147e07
	github.com/toolkits/http v0.0.0-20150609122824-f3ac6e6c24be
	github.com/toolkits/net v0.0.0-20160910085801-3f39ab6fe3ce
	github.com/toolkits/nux v0.0.0-20200401110743-debb3829764a
	github.com/toolkits/proc v0.0.0-20170520054645-8c734d0eb018
	github.com/toolkits/slice v0.0.0-20141116085117-e44a80af2484
	github.com/toolkits/str v0.0.0-20160913030958-f82e0f0498cb
	github.com/toolkits/sys v0.0.0-20170615103026-1f33b217ffaf
	github.com/toolkits/time v0.0.0-20160524122720-c274716e8d7f
	github.com/wvanbergen/kafka v0.0.0-20171203153745-e2edea948ddf
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a
	github.com/zloylos/grsync v1.7.0
	golang.org/x/sys v0.27.0
)

require (
	github.com/IBM/sarama v0.0.0-00010101000000-000000000000 // indirect
	github.com/Knetic/govaluate v3.0.0+incompatible // indirect
	github.com/alicebob/gopher-json v0.0.0-20180125190556-5a6b3ba71ee6 // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible // indirect
	github.com/beego/goyaml2 v0.0.0-20130207012346-5545475820dd // indirect
	github.com/beego/x2j v0.0.0-20131220205130-a0352aadc542 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20180710155616-bc664df96737 // indirect
	github.com/casbin/casbin v1.7.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudflare/golz4 v0.0.0-20150217214814-ef862a3cdc58 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/couchbase/go-couchbase v0.0.0-20200519150804-63f3cdb75e0d // indirect
	github.com/couchbase/gomemcached v0.0.0-20200526233749-ec430f949808 // indirect
	github.com/couchbase/goutils v0.0.0-20180530154633-e865a1461c8a // indirect
	github.com/cupcake/rdb v0.0.0-20161107195141-43ba34106c76 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/edsrzf/mmap-go v0.0.0-20170320065105-0bce6a688712 // indirect
	github.com/elastic/go-elasticsearch/v6 v6.8.5 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0 // indirect
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/glendc/gopher-json v0.0.0-20170414221815-dc4743023d0c // indirect
	github.com/go-redis/redis v6.14.2+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo/redis v0.0.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/niean/go-metrics-lite v0.0.0-20151230091537-b5d30971b578 // indirect
	github.com/niean/gotools v0.0.0-20151221085310-ff3f51fc5c60 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/peterh/liner v1.0.1-0.20171122030339-3681c2a91233 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/siddontang/go v0.0.0-20170517070808-cb568a3e5cc0 // indirect
	github.com/siddontang/goredis v0.0.0-20150324035039-760763f78400 // indirect
	github.com/siddontang/rdb v0.0.0-20150307021120-fc89ed2e418d // indirect
	github.com/signmem/viper v1.13.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ssdb/gossdb v0.0.0-20180723034631-88f6b59b84ec // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/syndtr/goleveldb v0.0.0-20181127023241-353a9fca669c // indirect
	github.com/tebeka/strftime v0.1.5 // indirect
	github.com/ugorji/go v0.0.0-20171122102828-84cb69a8af83 // indirect
	github.com/wendal/errors v0.0.0-20130201093226-f66c77a7882b // indirect
	github.com/yuin/gopher-lua v0.0.0-20171031051903-609c9cd26973 // indirect
	golang.org/x/text v0.15.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
