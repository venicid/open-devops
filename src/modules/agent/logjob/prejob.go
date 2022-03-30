package logjob

import (
	"crypto/md5"
	"encoding/hex"
	"open-devops/src/modules/agent/config"
	"open-devops/src/modules/agent/reader"
)

type LogJob struct {
	//r string  // 代表我们的生产者
	r *reader.Reader // 读取日志

	c string  // 代表我们的消费者
	Stra *config.LogStrategy  // 策略
}

func (lj *LogJob) hash() string  {
	md5obj := md5.New()
	md5obj.Write([]byte(lj.Stra.MetricName))
	md5obj.Write([]byte(lj.Stra.FilePath))
	return hex.EncodeToString(md5obj.Sum(nil))
}

func (lj *LogJob) start()  {
	
}

func (lj *LogJob) stop()  {

}