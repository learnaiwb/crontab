package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main()  {
	var (
		client *clientv3.Client
		err error
		config clientv3.Config
		ctx context.Context
		cancelFunc context.CancelFunc
		resp *clientv3.GetResponse
	)
	config = clientv3.Config{
		Endpoints: []string{"123.57.194.18:2379"},
		DialTimeout: 5 * time.Second,
	}
	if client, err = clientv3.New(config);err != nil{
		fmt.Println(err)
		return
	}
	ctx, cancelFunc = context.WithTimeout(context.TODO(),5 * time.Second)
	defer cancelFunc()
	//get
	if resp, err = client.Get(ctx,"/cron/jobs",clientv3.WithPrefix());err != nil{
		fmt.Println(err)
		return
	}

	for k, v := range resp.Kvs{
		fmt.Println(k,v)
	}
}
