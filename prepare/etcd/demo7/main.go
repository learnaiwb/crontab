package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

//分布式锁实现

func main() {
	var (
		config           clientv3.Config
		client           *clientv3.Client
		lease            clientv3.Lease
		leaseGrantResp   *clientv3.LeaseGrantResponse
		leaseKeepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
		leaseId          clientv3.LeaseID
		txn              clientv3.Txn
		txnResp          *clientv3.TxnResponse
		ctx              context.Context
		cancelFunc       context.CancelFunc

		err error
	)
	config = clientv3.Config{
		Endpoints:   []string{"*.*.*.*:2379"},
		DialTimeout: 5 * time.Second,
	}
	//1 建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	//2 上锁 创建lease 自动续约
	lease = clientv3.NewLease(client)

	if leaseGrantResp, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println(err)
		return
	}
	leaseId = leaseGrantResp.ID
	// 自动续约
	ctx, cancelFunc = context.WithCancel(context.TODO())
	defer cancelFunc()
	defer lease.Revoke(context.TODO(), leaseId)

	if leaseKeepAliveCh, err = lease.KeepAlive(ctx, leaseId); err != nil {
		fmt.Println(err)
		return
	}
	//处理续约应答的协程
	go func() {
		for {
			select {
			case resp, ok := <-leaseKeepAliveCh:
				if !ok {
					fmt.Println("通道关闭")
					return
				}
				fmt.Println("收到自动续租应答 ttl: ", resp.TTL)
			}
		}
	}()

	//抢锁 事务
	txn = client.Txn(context.TODO())
	txnResp, err = txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job1"), "=", 0)).
		Then(clientv3.OpPut("/cron/lock/job1", "get11", clientv3.WithLease(leaseId)),
			clientv3.OpPut("/cron/lock/job1", "get12", clientv3.WithLease(leaseId)),
		).
		Else(clientv3.OpGet("/cron/lock/job1")).Commit()
	if err != nil {
		fmt.Println(err)
		return
	}
	//判断是否抢到了锁
	if !txnResp.Succeeded {
		fmt.Println("锁被占用了", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	//3 处理业务
	fmt.Println("处理业务")
	time.Sleep(15 * time.Second)
	//4 释放锁
	//已经通过defer函数确保锁的释放
}
