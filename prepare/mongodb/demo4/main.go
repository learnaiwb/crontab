package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func main() {
	var (
		client     *mongo.Client
		collection *mongo.Collection
		err        error
	)
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://*.*.*.*:27017"))
	if err != nil {
		log.Println(err)
		return
	}

	collection = client.Database("my_db").Collection("cron")

	//构造删除条件
	//{"time_point.start_time":{"$lt":timestamp}}
	type TimeBefore struct {
		Before int64 `bson:"$lt"`
	}
	type DeleteCond struct {
		beforeCond TimeBefore `bson:"time_point.start_time"`
	}

	delCond := &DeleteCond{beforeCond: TimeBefore{Before: time.Now().Unix()}}

	res, err := collection.DeleteMany(context.TODO(), delCond)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(res.DeletedCount)
}
