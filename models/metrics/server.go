package metrics

import (
	"fmt"

	"github.com/gin-gonic/gin"
	logs "github.com/souliot/siot-log"
)

type MetricsServer struct {
	Port int
}

func NewMetricsServer(port int) (m *MetricsServer) {
	return &MetricsServer{
		Port: port,
	}
}

func (s *MetricsServer) Start() {
	DefaultMetrics.Init()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	InitRouter(r)
	go r.Run(fmt.Sprintf(":%d", s.Port))
	logs.Info("开启服务监控，端口：%d", s.Port)
}
