package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"time"
)

//https://github.com/etcd-io/etcd/tree/main/client/v3
func main()  {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"123.57.194.18:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
		fmt.Println(err)
		return
	}
	defer cli.Close()

	//put 操作
	ctx ,cancelFunc := context.WithTimeout(context.TODO(),10 * time.Second)
	defer cancelFunc()
	kv := clientv3.NewKV(cli)
	if putRes,err :=  kv.Put(ctx,"/cron/jobs/2","hello"); err != nil{
		fmt.Println(err)
		return
	}else {
		fmt.Println(putRes.Header.Revision)
		if putRes.PrevKv != nil {
			fmt.Println( string(putRes.PrevKv.Value))
		}
	}
}
