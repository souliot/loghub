package internal

import (
	"fmt"
	"loghub/config"
	"loghub/models/logcollect"
	"loghub/models/metrics"
	"loghub/models/srv"
	"os"
	"path"
	"public/libs_go/syslib"
	"strings"
	"time"

	"github.com/souliot/gateway/master"
	"github.com/souliot/gateway/service"
	logs "github.com/souliot/siot-log"
)

type Server struct {
	cfg *config.ServerCfg
	ser *srv.RegService
	lc  *logcollect.LogCollect
	ms  *metrics.MetricsServer
	ldb *logcollect.LogDb
}

func NewServer(ops ...config.Option) (m *Server, err error) {
	config.InitConfig()
	cfg := config.DefaultServerCfg
	cfg.Apply(ops)
	config.InitLog(cfg)
	logs.Info("初始化服务配置...")
	logs.Info("服务版本号：", cfg.Version)

	// log collector
	ps := loadConf(cfg.AppName)
	cfg.Collector.Paths = append(cfg.Collector.Paths, ps...)
	lc := logcollect.NewLogCollect(cfg.Collector.Paths, time.Duration(cfg.Collector.Interval)*time.Second, cfg.GoPoolSize, cfg.LocalIP)

	// service metrics
	ms := metrics.NewMetricsServer(cfg.HttpPort)

	// service register
	id := cfg.Id
	if id == "" {
		id = master.GetID()
		cfg.Id = id
		config.Config.Set("Id", id)
		if err = cfg.SaveConfigFile(); err != nil {
			return
		}
	}
	addr := fmt.Sprintf("%s:%d", cfg.LocalIP, cfg.HttpPort)
	meta := &master.ServiceMeta{
		Id:             id,
		Path:           service.DefaultRegion,
		Typ:            cfg.AppName,
		Address:        addr,
		Version:        cfg.Version,
		MetricsType:    master.MetricsTypeSystem,
		MetricsAddress: fmt.Sprintf("%s", addr),
	}
	ser, err := srv.NewRegService(cfg.EtcdEndpoints, meta)
	if err != nil {
		return
	}

	m = &Server{
		cfg: cfg,
		lc:  lc,
		ms:  ms,
		ser: ser,
		ldb: logcollect.DefaultLogDb,
	}
	return
}

func (s *Server) Start() {
	// 全局配置
	srv.WatchGlobalSetting(s.cfg.EtcdEndpoints)
	// 初始化 clickhouse 数据库
	addr := ""
	if len(srv.GlobalSetting.ClickAddress) != 0 {
		addr = srv.GlobalSetting.ClickAddress
	}
	if len(addr) <= 0 {
		logs.Error("Clickhouse地址配置错误：地址为空")
		os.Exit(0)
		return
	}
	s.ldb.Init(addr, logcollect.WithDb(s.cfg.Collector.DBName), logcollect.WithTable(s.cfg.Collector.TableName))
	logs.Info("初始化数据库配置，Clickhouse地址：", addr)
	// 启动日志采集
	if s.lc != nil {
		s.lc.Start()
	}
	// 启动监控
	if s.ms != nil {
		s.ms.Start()
	}
	// 启动服务注册
	if s.ser != nil {
		s.ser.Start()
	}
	logs.Info("服务启动成功！")
}

func (s *Server) Stop() {
	if s.lc != nil {
		s.lc.Stop()
	}
	if s.ser != nil {
		s.ser.Stop()
	}
	logs.Info("服务关闭成功！")
}

func loadConf(cur string) (ps []string) {
	logs.Info("加载config.conf配置文件")
	ps = make([]string, 0)
	ps_cache := make(map[string]bool)
	config := new(syslib.Config)
	if err := config.LoadConfig("../config.conf"); err == nil {
		apps := config.GetValueSliceString("applications")
		for _, v := range apps {
			if strings.Contains(v, cur) {
				continue
			}
			p := path.Join("..", path.Dir(v), "logs")
			ps_cache[p] = true
		}
	}
	if err := config.LoadConfig("../daemon/config.conf"); err == nil {
		apps := config.GetValueSliceString("applications")
		for _, v := range apps {
			if strings.Contains(v, cur) {
				continue
			}
			p := path.Join(path.Dir(v), "logs")
			ps_cache[p] = true
		}
	}

	for k, _ := range ps_cache {
		ps = append(ps, k)
	}
	return
}
