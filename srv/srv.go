package srv

import (
	"loghub/config"
	"loghub/models/logcollect"
	"loghub/models/metrics"
	"os"
	"path"
	e "public/entities"
	"public/libs_go/gateway/master"
	"public/libs_go/servicelib"
	conf "public/libs_go/servicelib/config"
	"public/libs_go/syslib"
	"strconv"
	"strings"
	"time"

	"public/libs_go/logs"
)

var (
	appName        = "loghub"
	version        = "5.1.2.0"
	DefaultService = NewService()
	DefaultConf    = NewConfig(DefaultService)
	globals        *e.ServerSetting
	serviceType    = strconv.Itoa(700)
)

type Service struct {
	cfg       *config.ServerCfg
	lc        *logcollect.LogCollect
	ms        *metrics.MetricsServer
	ldb       *logcollect.LogDb
	clickaddr string
}

func NewService(ops ...config.Option) (m *Service) {
	config.InitConfig()
	cfg := config.DefaultServerCfg
	cfg.Apply(ops)
	cfg.AppName = appName
	config.InitLog(cfg)
	logs.Info("初始化服务配置...")
	logs.Info("服务版本号：", version)

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
		cfg.SaveConfigFile()
	}

	m = &Service{
		cfg: cfg,
		lc:  lc,
		ms:  ms,
		ldb: logcollect.DefaultLogDb,
	}
	return
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

func (s *Service) NodeId() (id string) {
	return s.cfg.Id
}

func (s *Service) SaveNodeId(id string) {
	return
}

func (s *Service) SerSetting(data []byte) *master.SerSetting {
	return nil
}

func (s *Service) AppSetting(data []byte) *master.AppSetting {
	return nil
}

func (s *Service) GlobalSetting() interface{} {
	return &e.ServerSetting{}
}

func (s *Service) OnGlobalSetting(c interface{}) {
	// 初始化 clickhouse 数据库
	globals = c.(*e.ServerSetting)
	if len(globals.ClickAddress) != 0 {
		s.clickaddr = globals.ClickAddress
	}
	if len(s.clickaddr) <= 0 {
		logs.Error("Clickhouse地址配置错误：地址为空")
		os.Exit(0)
		return
	}
	return
}

func (s *Service) Metrics() (data []byte) {
	return make([]byte, 0)
}

func (s *Service) Ext() (data interface{}) {
	return
}

func (s *Service) OnVersion(data []byte) {
	// data 为版本实体的序列化数据
	moduleVersion.CheckVersion(data)
	return
}
func (s *Service) OnController(data *master.ControllerValue) {
	// data 为控制命令
	close()
	return
}

func (s *Service) Start(c *conf.Setting) (port int, err error) {
	s.ldb.Init(s.clickaddr, logcollect.WithDb(s.cfg.Collector.DBName), logcollect.WithTable(s.cfg.Collector.TableName))
	logs.Info("初始化数据库配置，Clickhouse地址：", s.clickaddr)
	// 启动日志采集
	if s.lc != nil {
		s.lc.Start()
	}
	// 启动监控
	if s.ms != nil {
		s.ms.Start()
	}
	port = s.cfg.HttpPort
	return
}

func (s *Service) Stop() {
	if s.lc != nil {
		s.lc.Stop()
	}

	logs.Info("服务关闭成功！")
	time.Sleep(300 * time.Millisecond)
	os.Exit(0)
	return
}

func NewConfig(ser *Service) *conf.Config {
	return &conf.Config{
		EtcdEndpoints: ser.cfg.EtcdEndpoints,
		ServiceType:   serviceType,
		MetricsType:   master.MetricsTypeNone,
		Version:       version,
	}
}

func close() {
	go servicelib.Stop(DefaultService)
	logs.Info("关闭程序")
	time.Sleep(300 * time.Millisecond)
	os.Exit(0)
}
