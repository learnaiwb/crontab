package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

type Job struct {
	Name string `json:"name"` //任务名
	Command string `json:"command"` //shell命令
	CronExpr string `json:"cronExpr"` //cron表达式
}
//任务调度计划
type JobSchedulePlan struct {
	Job *Job //要调度的任务信息
	Expr *cronexpr.Expression //解析好的cronExpr表达式
	NextTime time.Time //下次调度时间
}
//任务的执行状态
type JobExecuteInfo struct {
	Job *Job
	PlanTime time.Time //理论上调度时间
	RealTime time.Time //实际的调度实际
	Ctx context.Context //任务command的context
	CancelFunc context.CancelFunc //用于取消command执行的cancel函数
}


type Response struct {
	Errno int `json:"errno"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

type JobEvent struct {
	EventType int //save delete
	Job *Job
}
//任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo //执行状态
	Output []byte //脚本输出
	Err error //脚本执行错原因
	StartTime time.Time //启动时间
	EndTime time.Time //结束时间
}
//任务执行日志 给Mongodb
type JobLog struct {
	JobName string `bson:"jobName"`
	Command string `bson:"command"`
	Err string `bson:"err"`
	Output string `bson:"output"`
	PlanTime int64 `bson:"planTime"`
	ScheduleTime int64 `bson:"scheduleTime"`
	StartTime int64 `bson:"startTime"`
	EndTime int64 `bson:"endTime"`
}

//日志批次
type LogBatch struct {
	Logs []interface{}
}


// BuildResponse 序列化应答方法
func BuildResponse(errno int, msg string,data interface{}) (resp []byte,err error) {
	var res Response
	res.Errno = errno
	res.Msg = msg
	res.Data = data
	resp,err = json.Marshal(res)
	return
}
//用于etcd 反序列化内容
func UnpackJob(value []byte) (*Job,error) {
	var (
		err error
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(value,job);err != nil {
		return nil, err
	}
	return job,nil
}
//从etcd的key中提取函数
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey,JOB_SAVE_DIR)
}
//从etcd的key中提取函数
func ExtractKillerName(jobKey string) string {
	return strings.TrimPrefix(jobKey,JOB_KILL_DIR)
}

func BuildJobEvent(eventType int ,job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job: job,
	}
}

func BuildJobSchedulePlan(job *Job) (*JobSchedulePlan,error) {
	var (
		jobSchedulePlan *JobSchedulePlan
		expr *cronexpr.Expression
		err error
	)
	if expr,err = cronexpr.Parse(job.CronExpr);err != nil {
		return nil,err
	}

	jobSchedulePlan = &JobSchedulePlan{
		Job: job,
		Expr: expr,
		NextTime: expr.Next(time.Now()),
	}
	return jobSchedulePlan,nil
}

//构造执行状态信息
func BuildJobExecuteInfo(plan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      plan.Job,
		PlanTime: plan.NextTime, //计算调度时间
		RealTime: time.Now(), //真是时间
	}
	jobExecuteInfo.Ctx,jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

