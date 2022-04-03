package consumer

import (
	"bytes"
	"github.com/toolkits/pkg/logger"
	"math"
	"open-devops/src/modules/agent/config"
	"sort"
	"strconv"
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
	CounterQueue chan *AnalysPoint
	// 统计的字段
	Analyzing bool // 正在分析日志
}

// 从consumer往计算部分推的point
type AnalysPoint struct {
	Value float64  // 可能是数字的正则结果
	MetricsName string // metrics name
	LogFunc string   // 计算方法, cnt/max/min
	SortLabelString string  // 标签排序的结果
	LabelMap map[string]string
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



	defer func() {
		if err := recover();err != nil{
			logger.Errorf("[analysis.panic][mark:%v][err:%v]", c.Mark, err)
		}
	}()
	var (
		patternReg = c.Stra.PatternReg
		value = math.NaN()
		vString string  // 非cnt的正则计数 数字分组
	)

	/**
	## 处理日志主正则
	- patternReg.FindStringSubmatch(line) 的结果v
	- len=0 说明 正则没匹配中，应该丢弃这行
	- len=1 说明 正则匹配中了，但是小括号分组没匹配到
	- len>1 说明 正则匹配中了，小括号分组也匹配到
	*/
	

	// 处理日志主正则
	v := patternReg.FindStringSubmatch(line)
	if len(v) == 0{
		// 正则没匹配中，应该丢弃这行
		return
	}
	logger.Infof("[mark:%v][line:%v][reg_res:%v]", c.Mark, line, v)
	if len(v) > 1{
		// 说明 正则匹配中了，但是小括号分组没匹配到
		vString = v[1]
	}
	// 如果value能被解析为float，说明配置的，正则分组，应该是code=200
	value ,_ = strconv.ParseFloat(vString, 64)

	// 处理tag的正则
	labelMap := map[string]string{}
	for key, regTag := range c.Stra.TagRegs {
		labelMap[key] = ""
		t := regTag.FindStringSubmatch(line)
		if t != nil && len(t) > 1{
			labelMap[key] = t[1]
		}
	}

	ret := &AnalysPoint{
		Value:           value,
		MetricsName:     c.Stra.MetricName,
		LogFunc:         c.Stra.Func,
		SortLabelString: SortedTags(labelMap),
		LabelMap:        labelMap,
	}

	c.CounterQueue <- ret





}



func SortedTags(tags map[string]string) string {
	if tags == nil {
		return ""
	}

	size := len(tags)
	if size == 0 {
		return ""
	}

	ret := new(bytes.Buffer)

	if size == 1 {
		for k, v := range tags {
			ret.WriteString(k)
			ret.WriteString("=")
			ret.WriteString(v)
		}
		return ret.String()
	}

	keys := make([]string, size)
	i := 0
	for k := range tags {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for j, key := range keys {
		ret.WriteString(key)
		ret.WriteString("=")
		ret.WriteString(tags[key])
		if j != size-1 {
			ret.WriteString(",")
		}
	}

	return ret.String()
}
