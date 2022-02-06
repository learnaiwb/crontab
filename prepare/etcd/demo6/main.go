package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func main() {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		op      clientv3.Op
		opResp  clientv3.OpResponse
		putResp *clientv3.PutResponse
		getResp *clientv3.GetResponse
		err     error
	)
	config = clientv3.Config{
		Endpoints:   []string{"*.*.*.*:2379"},
		DialTimeout: 5 * time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		log.Println(err)
		return
	}
	op = clientv3.OpPut("/cron/test/1", "1")
	if opResp, err = client.Do(context.TODO(), op); err != nil {
		log.Println(err)
		return
	}
	putResp = opResp.Put()
	fmt.Println("put revison", putResp.Header.Revision)

	op = clientv3.OpGet("/cron/test/1")

	if opResp, err = client.Do(context.TODO(), op); err != nil {
		log.Println(err)
		return
	}

	getResp = opResp.Get()
	fmt.Println(getResp.Kvs[0].ModRevision)
	fmt.Println(string(getResp.Kvs[0].Key), string(getResp.Kvs[0].Value))

}
