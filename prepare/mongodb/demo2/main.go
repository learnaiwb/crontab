package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
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
		client       *mongo.Client
		collection   *mongo.Collection
		insertOneRes *mongo.InsertOneResult

		docID primitive.ObjectID

		err error
	)
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://*.*.*.*:27017"))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Disconnect(context.TODO())
	//选择数据库和表
	collection = client.Database("my_db").Collection("cron")

	record := &LogRecord{
		JobName: "job1",
		Command: "echo hello",
		Err:     "",
		Content: "hello",
		TimePoint: TimePoint{
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + 10,
		},
	}
	record = record
	record1 := bson.M{
		"JobName": "job1",
		"Command": "echo hello",
		"Err":     "",
		"Content": "hello",
		"TimePoint": bson.M{
			"StartTime": time.Now().Unix(),
			"EndTime":   time.Now().Unix() + 10,
		},
	}

	insertOneRes, err = collection.InsertOne(context.TODO(), record1)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(insertOneRes.InsertedID)
	docID = insertOneRes.InsertedID.(primitive.ObjectID)
	fmt.Println("docId", docID.Hex())
}
