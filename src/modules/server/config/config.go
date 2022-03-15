package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// 定义Mysql多个库
type Config struct {
	MysqlS []*MySQLConf `yaml:"mysql_s"`
	RpcAddr string `yaml:"rpc_addr"`
	HttpAddr string `yaml:"http_addr"`
	PCC PublicCloudSyncConf `yaml:"public_cloud_sync"`
}

type PublicCloudSyncConf struct {
	Enable bool `yaml:"enable"`

}


// 定义mysql单一库
type MySQLConf struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
	Max int `yaml:"max"`  // 连接数
	Idle int `yaml:"idle"`
	Debug bool `yaml:"debug"` // xorm打印sql

}

// 根据io read读取配置文件后的字符串，解析yaml
func Load(s []byte) (*Config, error) {
	cfg := &Config{}

	err := yaml.Unmarshal(s, cfg)
	if err != nil{
		return nil,  err
	}
	return cfg, nil
}

// 根据conf路径读取内容
func LoadFile(filename string) (*Config, error)  {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := Load(content)
	if err != nil{
		//fmt.Println("[parsing Yaml file err ...][err:%v]\n", err)
		return nil, err
	}
	return cfg, nil
}