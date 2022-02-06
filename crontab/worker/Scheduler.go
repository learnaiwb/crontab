package worker

import (
	"github.com/learnaiwb/crontab/crontab/common"
	"time"
)

type Scheduler struct {
	jobEventChan chan *common.JobEvent //etcd事件任务队列
	jobPlanTable map[string]*common.JobSchedulePlan //任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo //正在执行任务信息
	jobResultChan chan *common.JobExecuteResult //任务执行结果队列
}
var (
	G_scheduler *Scheduler
)

func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent)  {
	scheduler.jobEventChan <- jobEvent
}



//重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule () (scheduleAfter time.Duration) {
	var(
		jobPlan *common.JobSchedulePlan
		now time.Time
		nearTime *time.Time
	)
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}

	//当前时间
	now = time.Now()
	//1.遍历所有任务
	for _,jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// 尝试执行任务
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now)//更新下次执行时间
		}
		//统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime){
			nearTime = &jobPlan.NextTime
		}
	}
	//下次调度时间间隔
	scheduleAfter = (*nearTime).Sub(now)
	return
}
//尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan)  {
	//调度和执行是两件事 执行的任务可能很久 1分钟调度60次 但只能执行1次 防止并发 如果任务正在执行 跳过本次调度
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting bool
	)
	if jobExecuteInfo,jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name];jobExecuting {
		return
	}
	//构建正在执行的状态信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)
	//保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo
	//执行任务
	G_executor.ExecutorJob(jobExecuteInfo)
}

//调度协程
func (scheduler *Scheduler ) scheduleLoop()  {
	var (
		jobEvent *common.JobEvent //etcd任务事件队列
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobRes *common.JobExecuteResult
	)

	//初始化一次(1s)
	scheduleAfter = scheduler.TrySchedule()
	//调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	for  {
		select {
		case jobEvent = <-scheduler.jobEventChan: //接收处理任务事件
		scheduler.handleJobEvent(jobEvent)
		case <- scheduleTimer.C://最近的任务到期
		case jobRes = <-scheduler.jobResultChan: //监听任务执行结果
		scheduler.handleJobResult(jobRes)
		}
		scheduleAfter = scheduler.TrySchedule()
		//重置时间间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}
//处理任务事件
func (scheduler *Scheduler) handleJobEvent (jobEvent *common.JobEvent)  {
	var (
		plan *common.JobSchedulePlan
		err error
		jobExists bool
		jobExecuteInfo *common.JobExecuteInfo
		jobIsExecuting bool
	)

	switch jobEvent.EventType {
	case common.JOB_SAVE_EVENT: //保存任务事件
		if plan,err = common.BuildJobSchedulePlan(jobEvent.Job);err != nil {
			return //解析失败
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = plan
	case common.JOB_DELETE_EVENT: //删除任务事件
		if plan,jobExists = scheduler.jobPlanTable[jobEvent.Job.Name];jobExists {
			delete(scheduler.jobPlanTable,jobEvent.Job.Name)
		}
	case common.JOB_KILL_EVENT://强杀事件
		//取消掉command执行
		if jobExecuteInfo,jobIsExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name];jobIsExecuting {
			jobExecuteInfo.CancelFunc()
		}
	}
}

//处理执行结果
func (scheduler Scheduler) handleJobResult(jobRes *common.JobExecuteResult)  {

	var (
		jobLog *common.JobLog
	)

	delete(scheduler.jobExecutingTable,jobRes.ExecuteInfo.Job.Name)

	if jobRes.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		jobLog = &common.JobLog{
			JobName:      jobRes.ExecuteInfo.Job.Name,
			Command:      jobRes.ExecuteInfo.Job.Command,
			Output:       string(jobRes.Output),
			PlanTime:     jobRes.ExecuteInfo.PlanTime.UnixMilli(),
			ScheduleTime: jobRes.ExecuteInfo.RealTime.UnixMilli(),
			StartTime:    jobRes.StartTime.UnixMilli(),
			EndTime:      jobRes.EndTime.UnixMilli(),
		}
		if jobRes.Err != nil {
			jobLog.Err = jobRes.Err.Error()
		}
	}
	G_logSink.WriteLog(jobLog)
}


// PushJobResult 执行结果回传
func (scheduler Scheduler) PushJobResult(jobResult *common.JobExecuteResult)  {
	scheduler.jobResultChan <- jobResult
}


// InitScheduler 初始化调度器
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent,1000),
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan: make(chan *common.JobExecuteResult,1000),
	}
	go G_scheduler.scheduleLoop()
	return
}
