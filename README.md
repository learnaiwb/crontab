#crontab 使用go语言，基于etcd 及mongodb搭建一套分布式任务分发系统，可实现由网页下发任务到master节点，master节点负责下发、展示、删除、强杀任务编排，同统一格式编码后保存到etcd,
#worker进程负责任务发现调度、分布式锁抢占，任务执行，日志记录等功能实现；并可实现水平扩展worker节点。
#master 节点配置
#master/main/master.json中etcd服务器地址需自己搭建，输入自己ip
#启动master节点后，可以通过http://localhost:8070 查看控制界面

#worker节点配置
#worker/main/worker.json中需输入自己的etcd及mongodb地址
#worker启动后，即可消费下发的分布式任务
