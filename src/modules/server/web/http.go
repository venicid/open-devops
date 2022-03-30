package web

import (
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"net/http"
	"time"
)
func StartGin(httpAddr string, logger log.Logger) error  {

	r := gin.New()

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	// Gin Web Framework Prometheus metrics exporter
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	r.Use(gin.Logger())

	m := make(map[string]interface{})
	m["logger"] = logger
	r.Use(ConfigMiddle(m))

	// 设置路由
	configRoutes(r)

	s := &http.Server{
		Addr:              httpAddr,
		Handler:           r,
		ReadTimeout:       time.Second * 5,
		WriteTimeout:      time.Second * 5,
		MaxHeaderBytes:    1 << 20,
	}

	level.Info(logger).Log("msg", "web_server_aviabled_at", "httpAddr", httpAddr)
	err := s.ListenAndServe()
	return err

}


/**

go gin prometheus 导出mertic

gin_request_size_bytes_sum 3100
gin_request_size_bytes_count 8
# HELP gin_requests_total How many HTTP requests processed, partitioned by status code and HTTP method.
# TYPE gin_requests_total counter
gin_requests_total{code="200",handler="open-devops/src/modules/server/web.GetLabelDistribution",host="localhost:8082",method="POST",url="/api/v1/resource-distribution?page_size=2000"} 4
gin_requests_total{code="200",handler="open-devops/src/modules/server/web.ResourceQuery",host="localhost:8082",method="POST",url="/api/v1/resource-query?page_size=2000"} 4
# HELP gin_response_size_bytes The HTTP response sizes in bytes.
# TYPE gin_response_size_bytes summary
gin_response_size_bytes_sum 14216
gin_response_size_bytes_count 8

 */