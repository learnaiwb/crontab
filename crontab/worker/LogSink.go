package worker

import (
	"context"
	"github.com/learnaiwb/crontab/crontab/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//mongodb存储日志
type LogSink struct {
	client *mongo.Client
	logCollection *mongo.Collection
	logChan chan *common.JobLog
}
var (
	G_logSink *LogSink
)
//批量写入日志
func (logSink LogSink) saveLogs(batch *common.LogBatch)  {
	logSink.logCollection.InsertMany(context.TODO(),batch.Logs)
}

func (logSink LogSink) writeLoop()  {
 	var (
		 log *common.JobLog
		 logBatch *common.LogBatch
		 timer *time.Timer
	)
	timer = time.NewTimer(1 * time.Second)
	logBatch = &common.LogBatch{}
	for  {
		select {
		case log = <- logSink.logChan:
			//写入到db
			logBatch.Logs = append(logBatch.Logs,log)
			if len(logBatch.Logs) >= G_Config.JobLogBatchSize {
				//if !timer.Stop() {
				//	<- timer.C
				//}
				logSink.saveLogs(logBatch)
				logBatch.Logs = logBatch.Logs[:0]
			}
		case <- timer.C:
			if len(logBatch.Logs) > 0 {
				logSink.saveLogs(logBatch)
				logBatch.Logs = logBatch.Logs[:0]
			}
			timer.Reset(1 * time.Second)
		}
	}
}
func (logSink LogSink) WriteLog(log *common.JobLog)  {
	select {
	case logSink.logChan<-log:
	default: //队列满了丢弃
	}
}


func InitLogSink() (err error) {
	var (
		client *mongo.Client
	)
	if client,err = mongo.Connect(context.TODO(),options.Client().ApplyURI(G_Config.MongoDBUri).SetConnectTimeout(time.Duration(G_Config.EtcdDialTimeout)*time.Millisecond));err != nil{
		return
	}
	G_logSink = &LogSink{
		client: client,
		logCollection: client.Database("cron").Collection("log"),
		logChan: make(chan *common.JobLog,1000),
	}
	go G_logSink.writeLoop()
		return
}