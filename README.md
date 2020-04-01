# <font color=Green size=5>**仅仅是为了记录Etcd的使用**</font>
```
Just Record Etcd used Record~
因为涉及到公司内部账号信息以及代码库信息，因此做了脱敏处理
```
# <font color=Green size=5>**号码交叉验证平台**</font>
* 说明
     本模块主要为号码交叉验证平台，主要从库中提取Top无结果的号码 然后通过云手机、模拟器以及破解api等手段从竞品中提取号码信息
   项目wiki：
   <a href="http://wiki.molen.com/pages/viewpage.action?pageId=1037027288" target="_blank">http://wiki.molen.com/pages/viewpage.action?pageId=1037027288</a>
     
   
* 运行
- master
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $BASEDIR/bin/greedy-master  $BASEDIR/master/main.go
./greedy-master &
```
- worker
```
待补充
```

* 日志查看
```
服务部署日志默认当前路径
master日志: tail -f ./log/greedy-master.log
worker日志: tail -f ./log/greedy-worker.log
```

* 数据相关
- mongo
```
mongo中主要存放所有的源数据 以及 处理后的号码数据 由世选@xushixuan@molen.com维护
database:hm_online
源数据表: hm_greedy_source   
结果表: hm_greedy_result
```
- mysql
```
mysql中主要存放与任务相关的号码信息，即待抓取数据 由服务端自己维护
测试库信息
  database: greedy
  user: rdswr
  pwd : rdswr2018
  host: 10.14.122.12
  port 3306
```



* 服务部署相关
```
测试环境 10.99.91.62:8585(仅master部分 第一期适配云手机仅仅使用master模块)
```

# <font color=Green size=5>**项目负责人**</font>
* 张锋<<a href="mailto:zhangfeng15@molen.com">zhangfeng15@molen.com</a>>