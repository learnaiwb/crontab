package main

import (
	"flag"
	"fmt"
	"github.com/learnaiwb/crontab/crontab/master"
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
	flag.StringVar(&confFile,"config","./master.json","传入master.json路径")
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
	if err = master.InitConfig(confFile);err != nil{
		goto ERR
	}
	//连接etcd
	if err = master.InitJobMgr();err != nil {
		goto ERR
	}
	//启动api http服务
	if err = master.InitApiServer();err != nil {
		goto ERR
	}
	<-forerver
	//正常退出
	return
	ERR:
		fmt.Println(err)
}
