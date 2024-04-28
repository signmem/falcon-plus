# log 说明

> edit by terry.zeng
> date: 2020-09-16

# 服务器控制说明  


已申请下面 DNS 解析  

|域名| ip 地址段 | DNS 记录|
|-----|------|------|
|falcon-plugin-sync001.vip.vip.com| 10.248.50.66 |A|
|falcon-plugin-sync002.vip.vip.com| 10.248.50.100 |A|
|falcon-plugin-sync003.vip.vip.com| 10.248.50.101 |A|
|falcon-plugin-sync004.vip.vip.com| 10.248.50.74 |A|
|falcon-plugin-sync005.vip.vip.com| 10.248.50.75 |A|
|falcon-plugin-sync006.vip.vip.com| 10.248.50.76 |A|
|falcon-plugin-sync007.vip.vip.com| 10.248.50.77 |A|


# 数据同步方法  

## plugin 提交标准  

> 1 git issue (清楚说明意图功能)  
> 2 git 验收  (git 上提供 grafana 验收方法)  
> 3 git merge (由管理员完成，否则直接回滚忽略)  
> 4 记录 merge commit id  

## janitors app 使用  

> 登录 janitors app 地址暂时只完成 linux 数据同步方法    
> 输入 commit id (唯一需要提供信息)  
> 选择上面所有服务器  
> 点击执行即可完成 git plugin 更新至 rsync server   


# DEBUG 提示  

服务器端信息

> rsync 配置文件存放  /apps/conf/rsyncd.conf   
> rsync 启动方法  /apps/sh/rsyncd.sh start    
> rsync 提供对外开放目录  rsync://servername/linux_read  (linux 专用) 只读功能  
> rsync 最新代码存放至 /apps/dat/linux_plugin (linux 专用)  
> rsync 提供对外开放目录  rsync://servername/windows_read  (windows 专用) 只读功能  
> rsync 最新代码存放至 /apps/dat/linux_plugin (windows 专用)  
