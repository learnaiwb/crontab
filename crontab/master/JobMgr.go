package master

import (
	"context"
	"encoding/json"
	"github.com/learnaiwb/crontab/crontab/common"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}
var G_jobMgr *JobMgr

func InitJobMgr() (err error) {
	var (
		client *clientv3.Client
		config clientv3.Config
		kv clientv3.KV
		lease clientv3.Lease
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

	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

func (jobMgr *JobMgr)SaveJob (newJob *common.Job) (oldJob *common.Job,err error) {
	var (
		key string
		bytes []byte
		putResp *clientv3.PutResponse
		job common.Job
	)
	key = common.JOB_SAVE_DIR + newJob.Name

	if bytes,err = json.Marshal(newJob);err != nil {
		goto ERR
	}
	if putResp,err = jobMgr.kv.Put(context.TODO(),key,string(bytes),clientv3.WithPrevKV());err != nil{
		goto ERR
	}
	if putResp.PrevKv == nil {
		return nil,nil
	}

	if err = json.Unmarshal(putResp.PrevKv.Value,&job);err != nil {
		goto ERR
	}
	return &job,nil

	ERR:
		return nil, err
}

func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *common.Job,err error) {
	var (
		deleteResp *clientv3.DeleteResponse
		job common.Job
	)
	key := common.JOB_SAVE_DIR + name
	if deleteResp,err = jobMgr.kv.Delete(context.TODO(),key,clientv3.WithPrevKV());err != nil {
		goto ERR
	}
	if deleteResp.Deleted <= 0 {
		return nil, nil
	}
	if err = json.Unmarshal(deleteResp.PrevKvs[0].Value,&job);err != nil {
		goto ERR
	}
	return &job,nil

	ERR:
		return nil, err
}

func (jobMgr JobMgr) ListJobs() ([]*common.Job,error) {
	var (
		getResp *clientv3.GetResponse
		jobs []*common.Job
		err error
	)
	key := common.JOB_SAVE_DIR

	if getResp,err = jobMgr.kv.Get(context.TODO(),key,clientv3.WithPrefix());err != nil {
		goto ERR
	}

	if getResp.Count <= 0 {
		return nil,nil
	}

	for _,v := range getResp.Kvs {
		var temp common.Job
		if err = json.Unmarshal(v.Value,&temp);err != nil {
			continue
		}
		jobs = append(jobs,&temp)
	}
	return jobs,nil
	ERR:
		return nil,err
}

func (jobMgr JobMgr) KillJob (name string) (err error) {
	var (
		key string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
	)
	if leaseGrantResp,err = jobMgr.lease.Grant(context.TODO(),5);err != nil {
		goto ERR
	}
	leaseId = leaseGrantResp.ID
	key = common.JOB_KILL_DIR + name
	if _,err = jobMgr.kv.Put(context.TODO(),key,"",clientv3.WithLease(leaseId));err != nil {
		goto ERR
	}
	return nil

	ERR:
		return err
}