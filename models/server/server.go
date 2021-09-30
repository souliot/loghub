package server

import (
	"fmt"

	"loghub/models/metrics"
	"loghub/models/ws"
	"public/libs_go/logs"

	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	Port int
}

func NewApiServer(port int) (m *ApiServer) {
	return &ApiServer{
		Port: port,
	}
}

func (s *ApiServer) Start() {
	metrics.DefaultMetrics.Init()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	InitRouter(r)
	ws.InitWS()
	go r.Run(fmt.Sprintf(":%d", s.Port))
	logs.Info("开启服务监控，端口：%d", s.Port)
}
