package master

import (
	"encoding/json"
	"errors"
	"github.com/learnaiwb/crontab/crontab/common"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	G_ApiServer *ApiServer
)

//任务Http接口
type ApiServer struct {
	httpserver *http.Server
}

func InitApiServer() (err error) {
	var (
		mux *http.ServeMux
		httpserver *http.Server
		listener net.Listener
		staticDir http.Dir
	)
	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save",handleJobSave)
	mux.HandleFunc("/job/delete",handleJobDelete)
	mux.HandleFunc("/job/list",handleJobList)
	mux.HandleFunc("/job/kill",handleJobKill)
	//静态文件提供
	staticDir = http.Dir(G_Config.StaticDir)
	mux.Handle("/",http.StripPrefix("/",http.FileServer(staticDir)))


	if listener,err = net.Listen("tcp",":"+strconv.Itoa(G_Config.ApiPort));err != nil {
		return
	}

	//创建一个http服务
	httpserver = &http.Server{
		ReadTimeout: time.Duration(G_Config.APIReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_Config.APIWriteTimeout) * time.Millisecond,
		Handler: mux,
	}
	G_ApiServer = &ApiServer{
		httpserver: httpserver,
	}
	//启动服务端
	go func() {
		if err := httpserver.Serve(listener);err != nil {
			log.Println(err)
			return
		}
	}()

	return
}


//保存任务接口
//post job = {"name":"job1","command":"echo hello","cronExpr":"* * * * *"}
func handleJobSave(res http.ResponseWriter,r *http.Request)  {
	var (
		err error
		postJob string
		job common.Job
		oldJob *common.Job
		bytes []byte
	)
	//1 解析post表单
	//if err = r.ParseMultipartForm();err != nil {
	//	goto ERR
	//}
	//2 获取表单中的job字段
	postJob = r.FormValue("job")
	//3 反序列化job
	if err = json.Unmarshal([]byte(postJob),&job);err != nil {
		goto ERR
	}
	//4 保存job
	if oldJob,err = G_jobMgr.SaveJob(&job);err != nil {
		goto ERR
	}
	//5 返回应答
	if bytes,err = common.BuildResponse(0,"success",oldJob);err == nil{
		res.Write(bytes)
		return
	}
	ERR:
		if bytes,err = common.BuildResponse(-1,err.Error(),nil);err == nil {
			res.Write(bytes)
		}

}

//删除任务 post /job/delete name=job1
func handleJobDelete(res http.ResponseWriter,req *http.Request)  {

	var (
		bytes []byte
		err error
	)

	name := req.FormValue("name")
	oldjob,err := G_jobMgr.DeleteJob(name)
	if err != nil {
		goto ERR
	}
	if bytes,err = common.BuildResponse(0,"success",oldjob);err != nil {
		goto ERR
	}
	res.Write(bytes)
	return
ERR:
	if bytes,err = common.BuildResponse(-1,err.Error(),nil);err != nil  {
		res.Write(bytes)
	}
}

//列举所有job任务
func handleJobList(res http.ResponseWriter, req *http.Request)  {
	var (
		jobRes []*common.Job
		err error
		resp []byte
	)
	if jobRes,err = G_jobMgr.ListJobs();err != nil {
		goto ERR
	}

	resp,_ = common.BuildResponse(0,"success",jobRes)
	res.Write(resp)
	return
	ERR:
		resp,_ = common.BuildResponse(-1,err.Error(),nil)
		res.Write(resp)
}

//杀死任务 post /job/kill name=job1
func handleJobKill(res http.ResponseWriter,req *http.Request)  {
	var (
		jobName string
		err error
		bytes []byte
	)
	jobName = req.FormValue("name")
	if jobName == "" {
		err = errors.New("name不可为空")
		goto ERR
	}
	if err = G_jobMgr.KillJob(jobName); err != nil {
		goto ERR
	}
	bytes,_ = common.BuildResponse(0,"success",nil)
	res.Write(bytes)
	return
	ERR:
		bytes,_ = common.BuildResponse(-1,err.Error(),nil)
		res.Write(bytes)
}