package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {

	EtcdEndPoints []string `json:"etcdEndpoints"`
	EtcdDialTimeout int `json:"etcdDialTimeout"`

	MongoDBUri string `json:"mongodbUri"`
	MongoDBConnectTimeout int `json:"mongodbConnectTimeout"`
	JobLogBatchSize int `json:"jobLogBatchSize"`
	ExecutorLocation string `json:"executor"`
}

var G_Config *Config
//加载配置
func InitConfig(filename string) (err error) {
	var (
		config  Config
		content []byte
	)
	if content,err = ioutil.ReadFile(filename);err != nil {
		return
	}

	if err = json.Unmarshal(content,&config);err != nil {
		return err
	}
	G_Config = &config

	return
}