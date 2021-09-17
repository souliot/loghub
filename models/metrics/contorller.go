package metrics

import (
	"github.com/gin-gonic/gin"
)

type MetricsController struct{}

func (c *MetricsController) Metrics(ctx *gin.Context) {
	Handler.ServeHTTP(ctx.Writer, ctx.Request)
}
