title pingcheck 原理

```
pingcheck -> redis: 访问/falcon.ping 及 /falcon.pingdie\n获取主机 (无法ping)\n删除过期 1 小时主机
pingcheck -> cmdb: 1mins\n获取物理机信息\n过滤无法ping主机
pingcheck -> pingproxy: 请求 ping ipaddress
pingproxy -> 物理机: ping 检测
pingproxy -> pingcheck: 返回成功/失败
pingcheck -> redis: 失败的主机 /falcon.ping/ipaddress
pingmain -> redis: 20s 获取失败主机
pingmain -> pigeon: 120s 过期主机\n收敛后告警
pingmain -> redis: 告警后删除 /falcon.ping 过期主机\n过期主机写入 /falcon.pingdie
pingmain -> pingproxy: ping ipaddress
pingproxy -> 物理机: ping 检测
pingproxy -> pingmain: 返回成功/失败
pingmain -> redis: ping 通删除, 不通不处理
```
