# 新增 falcon agent 功能说明  

## plugin 自动同步  

> 已取消 falcon 使用 git 进行 plugin 更新方法  
> 利用 rsync 进行数据更新  
> agent 启动后，会获取 cfg.json 中 "rsyncaccess" 信息 ex: http://falcon-plugin.vip.vip.com  (nginx)  
> nginx 返回下面信息  
> agent 每天会随机选择一个时间点进行时间同步  
> 自动同步限制如下  
>> 1 nginx 设定为 sync : false 不可同步   
>> 2 IP 地址不匹配 source 范围不可同步  
>> 3 agent plugin 版本与 nginx 返回 version 版本一致不可同步  
> 
> 注意: 重启 vip-falcon-agent 会完成自动增量同步 plugin 一次不受上述限制  


格式参考  

```
{
  "syncdelete": false,
  "rpmupdate": true,
  "sync": true,
  "server": [
    "ns-yun-020022.vclound.com",
    "ns-yun-020023.vclound.com",
    "ns-yun-020024.vclound.com"
  ],
  "source": "0.0.0.0/0",
  "version": "20200927095310",
  "syncdest": "/apps/svr/falcon-agent/plugin/",
  "syncsrc": "/linux_sync",
  "rpmversion": {
    "el8": "http://mirrors.vclound.com/vclound/tmp/falcon/8/x86_64/vip-falcon-agent-1.0.2-7.el8.x86_64.rpm",
    "el6": "http://mirrors.vclound.com/vclound/tmp/falcon/6/x86_64/vip-falcon-agent-1.0.2-7.el6.x86_64.rpm",
    "el7": "http://mirrors.vclound.com/vclound/tmp/falcon/7/x86_64/vip-falcon-agent-1.0.2-7.el7.x86_64.rpm"
  },
  "port": "873"
}
```
## nginx 说明  


| 变量 |  类型 | 说明 |  
|----|----|----|  
| source | string | 只支持带 /mask 使用， ex: 0.0.0.0/0, 10.1.1.0/24, 10.2.0.0/16 暂时只作为 string 定义， 不支持 list |  
| sync| false| false 不需要同步， true 需要同步 用于控制 falcon plugin 更新   |  
| server| list | ['server1', 'server2', 'server3']  由 client 随机选取 list  |   
| syncdest| string |  需要同步的目录位置，避免用户修改配置文件导致问题 |  
| syncsrc| string | rsync 服务器位置 |  
| port|  string | rsync server port |  
| syncdelete| bool | 使用镜像同步， 默认 false   |  
| version| string |"yyyymmddhhmmss" 版本信息用于校验 agent 版本与 server 之间差异 |  
| rpmupdate | bool | 用于控制 rpm 版本更新 , true 更新 RPM， false 不更新 rpm |
| rpmversion|  dict  | rpm 版本下载地址, 当前只提供 el6, el7 |  


## 本机 plugin 查询方法

参考方法

> 直接读 /apps/svr/falcon-agent/plguin/version  
> 执行 falcon-agent -v  
> 执行 curl http://127.0.0.1:22230/version    

```
curl http://localhost:22230/version
{"msg":"success","data":{"falcon-agent":"5.1.6_20200909","falcon-plugin":"10-32@20200908144442"}}

10-32 为本机执行 cron job 同步 plugin 时间即 10:32 每天执行
20200908144442 plugin version    
```
特别注意:

```
当无法访问到 nginx 服务器，则版本信息如下 (方便从 db 中获取异常主机而设计)  
同样，假如 nginx 维护问题，格式按上述规格填写 json 格式，也会返回下面信息  
{"msg":"success","data":{"falcon-agent":"5.1.6_20200909","falcon-plugin":"16-44@plugin_http_access_error"}}
```

# agent cronjob 更新

> 默认会随机选择 10am - 18pm 一个时间点执行 rsync plugin 同步请求 (restful get 方法)    
>> 查询本机 rsync 更新时间方法, 参考 example 说明当前会在 15:36 执行同步    

```
curl http://localhost:22230/cron/gettime
{"msg":"success","data":{"hour":"15","minite":"36"}} (返回成功信息)
```

> 假如需要随机选择另外一个时间 (restful get 方法)  

```
curl http://localhost:22230/cron/reset
{"msg":"success","data":{"hour":"16","minite":"27"}}  (返回成功信息)  
```

> 假如需要手动选择一个时间 (restful post 方法)   
>> 时间必须选择在 10am - 18pm 以内 [合集]

```
curl -H "Accept: application/json" -H "Content-type: application/json" -X POST -d '{"hour":10, "minite":32}' http://localhost:22230/cron/settime
{"msg":"success","data":{"hour":"10","minite":"32"}} (返回成功信息)

```

# agent 手动执行 plugin 更新  

> 手动同步与自动同步区别  
>> 手动同步不受 nginx  sync : true, false 限制    
>> 手动同步不受 nginx  source 地址范围限制  
>> 手动同步不受 version 版本相同限制  
>
> 增量同步方法    

```
curl http://127.0.0.1:22230/plugin/update
success

``` 
> 镜像同步方法  

```
curl http://127.0.0.1:22230/plugin/reset
success
```

# 异常:  
返回下面信息都表明 agent 无法连接到 nginx 或无法从 nginx 获取正确信息  

```
curl http://localhost:22230/version
{"msg":"success","data":{"falcon-agent":"5.1.6_20200909","falcon-plugin":"16-44@plugin_http_access_error"}}

curl http://localhocurl http://localhost:22230/plugin/reset
plugin_http_access_error

curl http://localhost:22230/plugin/update
plugin_http_access_error

```
