package web

import (
	"github.com/gin-gonic/gin"
	"time"
)

func configRoutes(r *gin.Engine)  {
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.String(200, "ping")
		})
		api.GET("now-ts", GetNowTs)
		api.POST("/node-path", NodePathAdd)
		api.GET("/node-path", NodePathQuery)
		api.POST("/resource-mount", ResourceMount)
		api.DELETE("/resource-unmount", ResourceUnMount)
		api.POST("/resource-query", ResourceQuery)
		api.GET("/resource-group", ResourceGroup)
		api.POST("/resource-distribution", GetLabelDistribution)
	}
}

func GetNowTs(c *gin.Context)   {
	c.String(200, time.Now().Format("2006-01-02 15:00:11"))
}
