package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
)

// 定义Mysql多个库
type Config struct {
	RpcServerAddr string `yaml:"rpc_server_addr"`


	// log的配置
	LogStrategies []*LogStrategy `yaml:"log_strategies"`
	HttpAddr string `json:"http_addr"`
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

// 根据io read读取配置文件后的字符串，解析yaml
func Load(s []byte) (*Config, error) {
	cfg := &Config{}

	err := yaml.Unmarshal(s, cfg)
	if err != nil{
		return nil,  err
	}

	// 加载log的配置
	cfg.LogStrategies = setLogRegs(cfg)
	return cfg, nil
}


/*

*/
// 用户配置的日志策略，可以是agent 本地的yaml，也可以是通过接口过来的
type LogStrategy struct {
	ID         int64             `json:"id" yaml:"-"`
	MetricName string            `json:"metric_name" yaml:"metric_name"`
	MetricHelp string            `json:"metric_help" yaml:"metric_help"`
	FilePath   string            `json:"file_path" yaml:"file_path"`
	Pattern    string            `json:"pattern" yaml:"pattern"`
	Func       string            `json:"func" yaml:"func"`
	Tags       map[string]string `json:"tags" yaml:"tags"`
	// 上面是yaml或者前端的配置

	PatternReg *regexp.Regexp            `json:"-" yaml:"-"` // 主正则
	TagRegs    map[string]*regexp.Regexp `json:"-" yaml:"-"` // 标签的正则

	Creator string `json:"creator"`
}

// 解析用户配置的日志策略正则
func setLogRegs(cfg *Config) []*LogStrategy {
	res := []*LogStrategy{}
	for _, st := range cfg.LogStrategies {
		st := st
		st.TagRegs = make(map[string]*regexp.Regexp)

		// 处理主正则
		if len(st.Pattern) != 0{
			reg, err := regexp.Compile(st.Pattern)
			if err != nil{
				fmt.Printf("compile pattern regexp failed:[name: %v][pat:%v][err:%v]\n",
					st.MetricName, st.Pattern,err)
				continue
			}
			st.PatternReg = reg
		}

		// 处理标签的正则
		for tagK, tagV := range st.Tags {
			reg, err := regexp.Compile(tagV)
			if err != nil{
				fmt.Printf("compile pattern regexp failed:[name: %v][pat:%v][err:%v]\n",
					st.MetricName, st.Pattern,err)
				continue
			}
			st.TagRegs[tagK] = reg

		}
		res = append(res, st)
	}
	return res
}