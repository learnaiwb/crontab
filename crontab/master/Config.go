package master

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiPort int `json:"apiPort"`
	APIReadTimeout int `json:"apiReadTimeout"`
	APIWriteTimeout int `json:"apiWriteTimeout"`

	EtcdEndPoints []string `json:"etcdEndpoints"`
	EtcdDialTimeout int `json:"etcdDialTimeout"`

	StaticDir string `json:"staticDir"`
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