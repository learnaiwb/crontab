package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func main()  {
	var (
		config clientv3.Config
		client *clientv3.Client
		leaseKeepAliveResp  <-chan *clientv3.LeaseKeepAliveResponse
		err error
	)
	config = clientv3.Config{
		Endpoints: []string{"123.57.194.18:2379"},
		DialTimeout: 5 * time.Second,
	}
	if client,err = clientv3.New(config);err != nil{
		fmt.Println(err)
		return
	}
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	leaseResp,err := lease.Grant(context.TODO(),5)

	if err != nil {
		log.Println(err)
		return
	}
	leaseId := leaseResp.ID

	ctx,cancelFunc := context.WithTimeout(context.TODO(),3 *time.Second)
	defer cancelFunc()
	leaseKeepAliveResp,err =  lease.KeepAlive(ctx,leaseId)
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		for {
			select {
			case resp,ok :=  <- leaseKeepAliveResp:
				if !ok {
					fmt.Println("通道关闭")
					return
				}
				fmt.Println(resp.String())
			}
		}
	}()

	putResp,err := kv.Put(context.TODO(),"china","中国",clientv3.WithLease(leaseResp.ID))
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(putResp.Header.Revision)

	for {
		getResp,err := kv.Get(context.TODO(),"china")
		if err != nil{
			log.Println(err)
			return
		}
		if getResp.Count > 0 {
			fmt.Println(getResp.Kvs)
		}else{
			fmt.Println("无结果值")
			break
		}
		time.Sleep(2 * time.Second)
	}
	cancelFunc()
	fmt.Println("another test")
	time.Sleep(1 * time.Second)
	putResp,err = kv.Put(context.TODO(),"test","1",clientv3.WithLease(leaseId))
	t := time.Now()
	for {
		getResp,err := kv.Get(context.TODO(),"test")
		if err != nil {
			log.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("到期了")
			break
		}
		fmt.Println("还未到期",getResp.Kvs)
	}
	d := time.Now().Sub(t).Seconds()
	fmt.Println(d)
}
