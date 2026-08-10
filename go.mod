go 1.25.0

module github.com/signmem/falcon-plus

replace github.com/ugorji/go v0.0.0-20171122102828-84cb69a8af83 => github.com/ugorji/go/codec v1.2.12

replace github.com/ugorji/go/codec v1.2.12 => github.com/ugorji/go v1.2.12

replace github.com/Shopify/sarama => github.com/IBM/sarama v1.43.3

replace github.com/IBM/sarama => github.com/Shopify/sarama v1.43.3

replace github.com/gomodule/redigo/redis v0.0.1 => github.com/gomodule/redigo v2.0.0+incompatible

replace github.com/soniah/gosnmp => github.com/gosnmp/gosnmp v1.22.0

require (
	github.com/astaxie/beego v1.10.1
	github.com/coreos/go-log v0.0.0-20180308165134-b22fd89e1882
	github.com/emirpasic/gods v1.18.1
	github.com/facebookgo/atomicfile v0.0.0-20151019160806-2de1f203e7d5
	github.com/garyburd/redigo v1.6.4
	github.com/gin-contrib/pprof v1.5.4
	github.com/gin-gonic/gin v1.12.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/gomodule/redigo v1.9.2
	github.com/gosnmp/gosnmp v1.44.0
	github.com/jinzhu/gorm v1.9.16
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/masato25/yaag v0.0.0-20170704095552-00862ec4db8e
	github.com/niean/goperfcounter v0.0.0-20160108100052-24860a8d3fac
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.70.1
	github.com/redis/go-redis/v9 v9.7.0
	github.com/satori/go.uuid v1.2.0
	github.com/signmem/VM-Detection v0.0.0-20241205085638-790402faf36c
	github.com/signmem/consistent v0.0.0-20240428121820-de78f9192db8
	github.com/signmem/go-log v0.0.0-20240403024112-86a3635680b9
	github.com/signmem/kafka v0.0.0-20241204035601-1e29a9cfcd08
	github.com/signmem/netlib v1.0.1
	github.com/signmem/redislock v0.0.0-20240429073207-03e83a3aa6eb
	github.com/signmem/sarama v0.0.0-20241204033045-f7fcfef3ca58
	github.com/signmem/sarama-cluster v0.0.0-20241204035012-182d8387677e
	github.com/signmem/slice v0.0.0-20241205063021-73247974625a
	github.com/signmem/sys v0.0.0-20241205030312-7a95166215f5
	github.com/signmem/viper v1.13.0
	github.com/sirupsen/logrus v1.9.3
	github.com/smartystreets/goconvey v1.8.1
	github.com/spf13/cobra v1.8.1
	github.com/tatsushid/go-fastping v0.0.0-20160109021039-d7bb493dee3e
	github.com/toolkits/cache v0.0.0-20190218093630-cfb07b7585e5
	github.com/toolkits/concurrent v0.0.0-20150624120057-a4371d70e3e3
	github.com/toolkits/conn_pool v0.0.0-20170512061817-2b758bec1177
	github.com/toolkits/container v0.0.0-20151219225805-ba7d73adeaca
	github.com/toolkits/core v0.0.0-20141116054942-0ebf14900fe2
	github.com/toolkits/cron v0.0.0-20150624115642-bebc2953afa6
	github.com/toolkits/file v0.0.0-20160325033739-a5b3c5147e07
	github.com/toolkits/http v0.0.0-20150609122824-f3ac6e6c24be
	github.com/toolkits/net v0.0.0-20160910085801-3f39ab6fe3ce
	github.com/toolkits/proc v0.0.0-20170520054645-8c734d0eb018
	github.com/toolkits/str v0.0.0-20160913030958-f82e0f0498cb
	github.com/toolkits/time v0.0.0-20160524122720-c274716e8d7f
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a
	github.com/zloylos/grsync v1.7.0
	golang.org/x/sys v0.47.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.15.0 // indirect
	github.com/bytedance/sonic/loader v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/niean/go-metrics-lite v0.0.0-20151230091537-b5d30971b578 // indirect
	github.com/niean/gotools v0.0.0-20151221085310-ff3f51fc5c60 // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.59.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/signmem/file v1.1.1 // indirect
	github.com/signmem/nux v1.2.1 // indirect
	github.com/smarty/assertions v1.15.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tebeka/strftime v0.1.5 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.mongodb.org/mongo-driver/v2 v2.5.0 // indirect
	golang.org/x/arch v0.22.0 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
