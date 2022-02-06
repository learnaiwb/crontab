package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		kv     clientv3.KV
	)

	config = clientv3.Config{
		Endpoints:   []string{"*.*.*.*:2379"},
		DialTimeout: 5 * time.Second,
	}
	client, err = clientv3.New(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	kv = clientv3.NewKV(client)
	watcher := clientv3.NewWatcher(client)
	go func() {
		for {
			kv.Put(context.TODO(), "/cron/jobs/1", "1")
			//kv.Put(context.TODO(),"/cron/jobs/1", "2")
			kv.Delete(context.TODO(), "/cron/jobs/1")
			time.Sleep(1 * time.Second)
		}

	}()

	getResp, err := kv.Get(context.TODO(), "/cron/jobs/1")
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(getResp.Header.Revision)
	revision := getResp.Header.Revision
	newRevision := revision + 1

	ctx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFunc()
	})

	watcherchan := watcher.Watch(ctx, "/cron/jobs/1", clientv3.WithPrevKV(), clientv3.WithRev(newRevision))
	for w := range watcherchan {
		for _, v := range w.Events {
			if v == nil {
				break
			}
			switch v.Type {
			case mvccpb.PUT:
				fmt.Println("PUT", w.Header.Revision, v.Kv.CreateRevision, v.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("Delete", w.Header.Revision, v.PrevKv.CreateRevision, v.PrevKv.ModRevision, v.Kv.CreateRevision, v.Kv.ModRevision)

			}
		}
	}

}
