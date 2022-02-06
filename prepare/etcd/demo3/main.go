package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main()  {
	config := clientv3.Config{
		Endpoints: []string{"123.57.194.18:2379"},
		DialTimeout: 5 * time.Second,
	}
	client,err := clientv3.New(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	deleteResp, err := client.Delete(context.TODO(),"/cron/jobs/",clientv3.WithPrefix(),clientv3.WithPrevKV())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(deleteResp.Deleted)
	fmt.Println(deleteResp.Header)
	for k,v := range deleteResp.PrevKvs {
		fmt.Println(k,string(v.Key),string(v.Value))
	}
}
