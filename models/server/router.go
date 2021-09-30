package server

import (
	"loghub/models/metrics"
	"loghub/models/ws"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	root := r.Group("")
	{
		root.GET("/metrics", (&metrics.MetricsController{}).Metrics)
	}
	// v1 version
	v1 := r.Group("/v1")
	{
		v1.GET("/logs", (&ws.WsController{}).Logs)
	}
}
