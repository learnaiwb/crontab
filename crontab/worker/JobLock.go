package worker

import (
	"context"
	"github.com/learnaiwb/crontab/crontab/common"
	clientv3 "go.etcd.io/etcd/client/v3"
)

//分布锁
type JobLock struct {
	kv clientv3.KV
	lease clientv3.Lease
	jobName string //任务名
	leaseId clientv3.LeaseID
	cancelFunc context.CancelFunc
	isLocked bool //是否上锁成功
}
//初始化锁
func InitJobLock(jobName string,kv clientv3.KV,lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv: kv,
		lease:lease,
		jobName:jobName,
	}
	return
}

func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		ctx context.Context
		cancelFunc context.CancelFunc
		leaseId clientv3.LeaseID
		leaseKeepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
		txn clientv3.Txn
		txnResp *clientv3.TxnResponse
		lockKey string //锁路径
	)
	//1 创建租约 5s
	if leaseGrantResp,err = jobLock.lease.Grant(context.TODO(),5);err != nil {
		return err
	}
	ctx,cancelFunc = context.WithCancel(context.TODO())

	leaseId = leaseGrantResp.ID
	//2 自动续租
	if leaseKeepAliveCh,err = jobLock.lease.KeepAlive(ctx,leaseId);err != nil {
		goto FAIL
	}
	go func() {
		var keepResp *clientv3.LeaseKeepAliveResponse
		for  {
			select {
			case keepResp = <- leaseKeepAliveCh:
				if keepResp == nil{
					 goto END
				}

			}
		}
		END:
	}()


	//3 创建事务
	txn = jobLock.kv.Txn(context.TODO())
	//4 事务抢锁
	lockKey = common.JOB_LOCK_DIR + jobLock.jobName
	//5 成功返回，失败释放续租；注意成功后执行完逻辑需释放
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey),"=",0)).
		Then(clientv3.OpPut(lockKey,"",clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))
	if txnResp,err = txn.Commit();err != nil {
		goto FAIL
	}
	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}
	jobLock.leaseId = leaseId //用于后续取消
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	return
	FAIL:
		cancelFunc()
		jobLock.lease.Revoke(context.TODO(),leaseId)
		return
}
func (jobLock *JobLock) UnLock()  {
	if jobLock.isLocked {
		jobLock.cancelFunc() //取消程序自动续约的协程
		jobLock.lease.Revoke(context.TODO(),jobLock.leaseId) //释放续约
	}
}