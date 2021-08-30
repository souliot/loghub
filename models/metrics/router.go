package metrics

import (
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	v1 := r.Group("/")
	v1.GET("/metrics", (&MetricsController{}).Metrics)
	v1.GET("/", (&MetricsController{}).Metrics)
}
