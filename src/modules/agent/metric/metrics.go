package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"open-devops/src/modules/agent/config"
)

func CreateMetrics(ss []*config.LogStrategy) map[string]*prometheus.GaugeVec {
	mmmap:= make(map[string]*prometheus.GaugeVec)
	for _, s := range ss {
		labels := []string{}

		for k := range s.Tags{
			labels =append(labels, k)
		}
		m := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: s.MetricName,
			Help: s.MetricHelp,
		}, labels)
		mmmap[s.MetricName] = m

	}
	return mmmap

}