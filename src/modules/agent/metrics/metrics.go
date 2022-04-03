package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/toolkits/pkg/logger"
	"net/http"
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

func StartMetricWeb(addr string) error  {
	logger.Infof("LogJobManager.StartMetricWeb.start: ", addr)
	http.Handle("/metrics", promhttp.Handler())
	srv := http.Server{
		Addr:              addr,
	}
	err := srv.ListenAndServe()
	return err
}