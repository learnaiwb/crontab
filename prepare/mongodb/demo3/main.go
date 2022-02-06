package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

//日志记录
type LogRecord struct {
	JobName   string    `bson:"job_name"` //任务名称
	Command   string    `bson:"command"`  //shell命令
	Err       string    `bson:"err"`      //脚本错误
	Content   string    `bson:"content"`  //脚本输出
	TimePoint TimePoint `bson:"time_point"`
}
type TimePoint struct {
	StartTime int64 `bson:"start_time"`
	EndTime   int64 `bson:"end_time"`
}

func main() {
	var (
		client     *mongo.Client
		collection *mongo.Collection
		err        error
	)
	// 1 建立连接
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://*.*.*.*:27017"))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Disconnect(context.TODO())

	//2 db collection
	collection = client.Database("my_db").Collection("cron")

	//3 find
	opts := options.Find().SetSkip(0).SetLimit(2)
	cursor, err := collection.Find(context.TODO(), bson.M{"jobname": "job1"}, opts)
	if err != nil {
		log.Println(err)
		return
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var record map[string]interface{}
		err = cursor.Decode(&record)
		if err != nil {
			log.Println(err)
			return
		}
		for k, v := range record {
			fmt.Println(k, v)
		}
	}
}
