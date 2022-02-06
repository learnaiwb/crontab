package worker

import (
	"context"
	"github.com/learnaiwb/crontab/crontab/common"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	watcher clientv3.Watcher
}
var G_jobMgr *JobMgr

func InitJobMgr() (err error) {
	var (
		client *clientv3.Client
		config clientv3.Config
		kv clientv3.KV
		lease clientv3.Lease
		watcher clientv3.Watcher
	)
	config = clientv3.Config{
		Endpoints:   G_Config.EtcdEndPoints,
		DialTimeout: time.Duration(G_Config.EtcdDialTimeout) * time.Millisecond,
	}
	if client,err = clientv3.New(config);err != nil {
		return
	}
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)
	//赋值单例
	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
		watcher: watcher,
	}
	//启动任务监听
	G_jobMgr.watchJobs()
	//启动监听killer
	G_jobMgr.watchKiller()
	return
}
//监听任务变化
func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp *clientv3.GetResponse
		kvpair *mvccpb.KeyValue
		job *common.Job
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		event *clientv3.Event
		deleteKey string
		jobEvent *common.JobEvent
	)
	//1 get一下/cron/jobs/目录下所有的任务，并且获知当前集群的revision
	if getResp,err = jobMgr.kv.Get(context.TODO(),common.JOB_SAVE_DIR,clientv3.WithPrefix());err != nil {
		goto ERR
	}
	//2 当前有哪些任务
	for _,kvpair = range getResp.Kvs {
		if job,err = common.UnpackJob(kvpair.Value);err == nil{
			jobEvent = common.BuildJobEvent(common.JOB_SAVE_EVENT,job)
			// send scheduler(调度协程)
			G_scheduler.PushJobEvent(jobEvent)
		}
	}
	//3 从revision开始监听
	go func() {
		watchStartRevision = getResp.Header.Revision + 1
		//监听目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(),common.JOB_SAVE_DIR,clientv3.WithRev(watchStartRevision),clientv3.WithPrefix())
		//处理监听事件
		for watchResp = range watchChan {
			for _,event = range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT: //保存任务事件
					if job,err = common.UnpackJob(event.Kv.Value);err != nil {
						continue
					}
					jobEvent = common.BuildJobEvent(common.JOB_SAVE_EVENT,job)
				case mvccpb.DELETE:
					deleteKey = common.ExtractJobName(string(event.Kv.Key))
					job = &common.Job{Name: deleteKey}
					jobEvent = common.BuildJobEvent(common.JOB_DELETE_EVENT,job)
				}
				// send scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()


	ERR:
		return err
}
//监听强杀任务通知
func (jobMgr *JobMgr) watchKiller()  {
	var (
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		event *clientv3.Event
		jobEvent *common.JobEvent
		jobName string
	)

	go func() {
		//监听目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(),common.JOB_KILL_DIR,clientv3.WithPrefix())
		//处理监听事件
		for watchResp = range watchChan {
			for _,event = range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT: //保存任务事件
					jobName = common.ExtractKillerName(string(event.Kv.Key))
					jobEvent = common.BuildJobEvent(common.JOB_KILL_EVENT,&common.Job{Name: jobName})
					// send scheduler
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: //标记任务自动过期
				}

			}
		}
	}()

	return
}

//创建任务执行锁
func (jobMgr *JobMgr)CreateJobLock(name string) (jobLock *JobLock) {
	//返回一把锁
	jobLock = InitJobLock(name,jobMgr.kv,jobMgr.lease)
	return
}