# 编译 agent 注意  

> 1 编译时候需要带 CGO_ENABLED=0  
> ( 建议使用 go1.14 编译)  

```
CGO_ENABLED=0 go build -o bin/agent/falcon-agent ./modules/agent
```
> 2 编译前确认 vendor


github.com/toolkits/sys/cmd.go  41 ~ 44   

```
//err = syscall.Setpgid(cmd.Process.Pid, cmd.Process.Pid)
//if err != nil {
//      log.Println("Setpgid failed, error:", err)
//}
```

> 3 确保 /usr/sbin/ss 文件位置正确 (centos5, 6, 7, 8) 已验证  
>> 在下面文件执行了硬编码， 否则 centos5 会报错 PortMetric()  

```
github.com/toolkits/nux/portstat.go  
```

# 编译 graph 注意   

>  graph 必须使用 centos6 系统执行编译否则无法在生产中正常运行
>  编译时候需要带 CGO_ENABLED=1  
>  否则会遇到交叉编译错误问题  ( 建议使用 go1.14 编译)  

```
CGO_ENABLED=1 go build -o bin/graph/falcon-graph ./modules/graph
```

# 编译 kafka_consumer 注意  
> 必须使用 go1.16 编译  


# log vendor 修改   

>  确保 vendor/github.com/coreos/go-log/log/fields.go 文件被修改  
> 1 修改 full_time 满足公司格式要求  

```
"full_time":  time.Now().Format("2006-01-02 15:04:05.999"),  // time of log entry

```

> 2 修改 logger.verbose = true  由于属于内部变量无法外部修改   

```
logger.verbose = true
```

> 3.修复 vendor/github.com/toolkits/sys/cmd.go +43 避免 Setpgid failed 报错

# graph db  

```
alter table
  endpoint_counter
add
  index `t_create_counter`(`t_create`, `counter`);
```
> 4.修复 vm，忽略所有错误信息

```
vendor/github.com/ShellCode33/VM-Detection/vmdetect/common.go  

删除错误输出  
PrintError
PrintWarning

vendor/github.com/ShellCode33/VM-Detection/vmdetect/linux.go
删除 Print 错误信息


``` 
> 5.修改挂载点问题  

open-falcon/falcon-plus/vendor/github.com/toolkits/nux/dfstat.go

```
var FSSPEC_IGNORE = map[string]struct{}{
	"none":      struct{}{},
	"nodev":     struct{}{},
	"proc":      struct{}{},
	"hugetlbfs": struct{}{},
	"mqueue":    struct{}{},
}

var FSTYPE_IGNORE = map[string]struct{}{
	"cgroup":     struct{}{},
	"debugfs":    struct{}{},
	"devpts":     struct{}{},
	"devtmpfs":   struct{}{},
	"iso9660":    struct{}{},
	"rpc_pipefs": struct{}{},
	"rootfs":     struct{}{},
	"overlay":    struct{}{},
	"tmpfs":      struct{}{},
	"squashfs":   struct{}{},
	"autofs":     struct{}{},
	"ceph":       struct{}{},
	"configfs":   struct{}{},
	"mqueue":     struct{}{},
	"proc":       struct{}{},
	"pstore":     struct{}{},
	"securityfs": struct{}{},
	"sysfs":      struct{}{},
	"tmpfs":      struct{}{},
}

var FSFILE_PREFIX_IGNORE = []string{
	"/sys",
	"/net",
	"/misc",
	"/proc",
	"/lib",
	"/run",
	"/mnt",
}
```



# Falcon+

![Open-Falcon](./logo.png)

[![Build Status](https://travis-ci.org/open-falcon/falcon-plus.svg?branch=plus-dev)](https://travis-ci.org/open-falcon/falcon-plus)
[![codecov](https://codecov.io/gh/open-falcon/falcon-plus/branch/plus-dev/graph/badge.svg)](https://codecov.io/gh/open-falcon/falcon-plus)
[![GoDoc](https://godoc.org/github.com/open-falcon/falcon-plus?status.svg)](https://godoc.org/github.com/open-falcon/falcon-plus)
[![Code Issues](https://www.quantifiedcode.com/api/v1/project/5035c017b02c4a4a807ebc4e9f153e6f/badge.svg)](https://www.quantifiedcode.com/app/project/5035c017b02c4a4a807ebc4e9f153e6f)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-falcon/falcon-plus)](https://goreportcard.com/report/github.com/open-falcon/falcon-plus)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)

# Documentations

- [Usage](http://book.open-falcon.org)
- [Open-Falcon API](http://open-falcon.org/falcon-plus)

# Prerequisite

- Git >= 1.7.5
- Go >= 1.6

# Getting Started

## Docker

Please refer to ./docker/[README.md](https://github.com/open-falcon/falcon-plus/blob/master/docker/README.md).

## Build from source
**before start, please make sure you prepared this:**

```
yum install -y redis
yum install -y mysql-server

```

*NOTE: be sure to check redis and mysql-server have successfully started.*

And then

```
# Please make sure that you have set `$GOPATH` and `$GOROOT` correctly.
# If you have not golang in your host, please follow [https://golang.org/doc/install] to install golang.

#mkdir -p $GOPATH/src/github.com/open-falcon
#cd $GOPATH/src/github.com/open-falcon
#git clone https://github.com/open-falcon/falcon-plus.git

mkdir -p $GOPATH/src/github.com
cd $GOPATH/src/github.com
git clone git@gitlab.tools.vipshop.com:vip-ops-sh/open-falcon.git

```

**And do not forget to init the database first (if you have not loaded the database schema before)**

```
cd $GOPATH/src/github.com/open-falcon/falcon-plus/scripts/mysql/db_schema/
mysql -h 127.0.0.1 -u root -p < 1_uic-db-schema.sql
mysql -h 127.0.0.1 -u root -p < 2_portal-db-schema.sql
mysql -h 127.0.0.1 -u root -p < 3_dashboard-db-schema.sql
mysql -h 127.0.0.1 -u root -p < 4_graph-db-schema.sql
mysql -h 127.0.0.1 -u root -p < 5_alarms-db-schema.sql
```

**NOTE: if you are upgrading from v0.1 to current version v0.2.0,then**. [More upgrading instruction](http://www.jianshu.com/p/6fb2c2b4d030)

    mysql -h 127.0.0.1 -u root -p < alarms-db-schema.sql

# Compilation

```
cd $GOPATH/src/github.com/open-falcon/falcon-plus/

# make all modules
make all

# make specified module
make agent

# pack all modules
make pack
```

* *after `make pack` you will got `open-falcon-vx.x.x.tar.gz`*
* *if you want to edit configure file for each module, you can edit `config/xxx.json` before you do `make pack`*

#  Unpack and Decompose

```
export WorkDir="$HOME/open-falcon"
mkdir -p $WorkDir
tar -xzvf open-falcon-vx.x.x.tar.gz -C $WorkDir
cd $WorkDir
```

# Start all modules in single host
```
cd $WorkDir
./open-falcon start

# check modules status
./open-falcon check

```

# Run More Open-Falcon Commands

for example:

```
# ./open-falcon [start|stop|restart|check|monitor|reload] module
./open-falcon start agent

./open-falcon check
        falcon-graph         UP           53007
          falcon-hbs         UP           53014
        falcon-judge         UP           53020
     falcon-transfer         UP           53026
       falcon-nodata         UP           53032
   falcon-aggregator         UP           53038
        falcon-agent         UP           53044
      falcon-gateway         UP           53050
          falcon-api         UP           53056
        falcon-alarm         UP           53063
```

* For debugging , You can check `$WorkDir/$moduleName/log/logs/xxx.log`

# Install Frontend Dashboard
- Follow [this](https://github.com/open-falcon/dashboard).

**NOTE: if you want to use grafana as the dashboard, please check [this](https://github.com/open-falcon/grafana-openfalcon-datasource).**

# Package Management

We use govendor to manage the golang packages. Please install `govendor` before compilation.

    go get -u github.com/kardianos/govendor

Most depended packages are saved under `./vendor` dir. If you want to add or update a package, just run `govendor fetch xxxx@commitID` or `govendor fetch xxxx@v1.x.x`, then you will find the package have been placed in `./vendor` correctly.

# Package Release

```
make clean all pack
```

# Q&A

Any issue or question is welcome, Please feel free to open [github issues](https://github.com/open-falcon/falcon-plus/issues) :)
