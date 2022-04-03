package logjob

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/toolkits/pkg/logger"
	"open-devops/src/common"
	"open-devops/src/modules/agent/config"
	"open-devops/src/modules/agent/consumer"
	"open-devops/src/modules/agent/reader"
)

type LogJob struct {
	//r string  // 代表我们的生产者
	r *reader.Reader // 读取日志

	//c string  // 代表我们的消费者
	c *consumer.ConsumerGroup // 代表我们的消费组组
	Stra *config.LogStrategy  // 策略
}

func (lj *LogJob) hash() string  {
	md5obj := md5.New()
	md5obj.Write([]byte(lj.Stra.MetricName))
	md5obj.Write([]byte(lj.Stra.FilePath))
	return hex.EncodeToString(md5obj.Sum(nil))
}

func (lj *LogJob) start(cq chan *consumer.AnalysPoint)  {

	logger.Infof("create.LogJob.start")

	fPath := lj.Stra.FilePath
	stream := make(chan string, common.LogQuerySize)
	// new reader
	r, err := reader.NewReader(fPath, stream)
	if err != nil{
		return
	}
	lj.r = r
	// new consumer
	cg := consumer.NewConsumerGroup(fPath, stream, lj.Stra, cq)
	lj.c = cg
	// 启动r ,c
	// 先消费者
	lj.c.Start()
	// 后生产者
	go r.Start()

	logger.Infof("[create.LogJob.success][fPath:%d][MetricName:%d]", fPath, lj.Stra.MetricName)



}

// 先停止生产者，后停止消费者
func (lj *LogJob) stop()  {

	lj.r.Stop()
	lj.c.Stop()

}