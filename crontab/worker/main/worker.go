package main

import (
	"flag"
	"fmt"
	"github.com/learnaiwb/crontab/crontab/worker"
	"runtime"
)

func initEnv()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var (
	confFile string //配置文件路径
)

//解析命令行参数
func initArgs()  {
	// main -config ./master.json
	// main -h
	flag.StringVar(&confFile,"config","./worker.json","worker.json路径")
	flag.Parse()
}

func main()  {

	var (
		err error
	)
	forerver := make(chan struct{})
	//初始化命令行参数
	initArgs()
	//初始化线程
	initEnv()
	//初始化配置
	if err = worker.InitConfig(confFile);err != nil{
		goto ERR
	}
	//启动日志协程
	if err = worker.InitLogSink();err != nil {
		goto ERR
	}

	//启动执行器
	if err = worker.InitExecutor();err != nil {
		goto ERR
	}

	//启动调度器
	if err = worker.InitScheduler();err != nil {
		goto ERR
	}


	//连接etcd
	if err = worker.InitJobMgr();err != nil {
		goto ERR
	}


	<-forerver
	//正常退出
	return
	ERR:
		fmt.Println(err)
}
