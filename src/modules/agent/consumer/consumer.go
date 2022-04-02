package consumer

import (
	"github.com/toolkits/pkg/logger"
	"open-devops/src/modules/agent/config"
	"time"
)

// 单个consumer对象
// TODO 缺少analyPoint
type Consumer struct {
	FilePath string
	Stream chan  string		// 接受生产者的chan
	Stra *config.LogStrategy  // 策略
	Mark string // work的名字，方便后续排查问题
	Close chan struct{}
	// 统计的字段
	Analyzing bool // 正在分析日志
}




func (c *Consumer) Start()  {
	go func() {
		c.work()
	}()
}

func (c *Consumer) Stop()  {
	close(c.Close)
}

func (c *Consumer) work()  {

	logger.Infof("[Consumer:%v]starting...{}", c.Mark)

	var anaCnt, anaSwp int64

	analyClose := make(chan struct{})

	go func() {
		select {
		case <- analyClose:
			return
		case <- time.After(10*time.Second):

		}
		a := anaCnt
		logger.Infof("[Consumer:%v][analysis %d line in last 10s]", c.Mark, a- anaSwp)
		anaSwp = a
	}()

	for  {
		select {
		case line := <- c.Stream:
			anaCnt ++
			c.Analyzing = true
			c.analysis(line)
			c.Analyzing = false
		case <- c.Close:
			analyClose <- struct{}{}
			return
		}
	}

}

func (c *Consumer) analysis(line string)  {
	logger.Infof("[Consumer:%v][analysis.line:%s]", c.Mark, line)

}
