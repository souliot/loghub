package controllers

import (
	"loghub/models"
	"loghub/models/config"
	"loghub/models/logcollect"
	"path"
	"public/libs_go/syslib"
	"time"

	slog "github.com/souliot/siot-log"
	"github.com/urfave/cli/v2"
)

var (
	LogHubController = new(LogHub)
	defaultClickAddr = "tcp://127.0.0.1:9008?username=default&password=watrix888"
	lc               *logcollect.LogCollect
)

type LogHub struct{}

func (c *LogHub) Run(ctx *cli.Context) {
	c.InitBase(ctx)
	log_paths := ctx.StringSlice("log_paths")
	ps := loadConf()
	log_paths = append(log_paths, ps...)
	slog.Info("日志采集目录：", log_paths)
	log_interval := ctx.Int64("log_interval")
	lc = logcollect.NewLogCollect(log_paths, time.Duration(log_interval)*time.Second, models.GoPoolSize)
	lc.Start()

	slog.Info("服务启动成功！")
}

func (c *LogHub) Stop() {
	if lc != nil {
		lc.Stop()
	}
}

func (c *LogHub) InitBase(ctx *cli.Context) {
	// 设置goroutine size
	models.GoPoolSize = ctx.Int("gopoolsize")
	models.TestMode = ctx.Bool("test")

	// ETCD
	etcdEndpoints := ctx.StringSlice("etcdendpoints")
	if len(etcdEndpoints) > 0 {
		config.WatchGlobalSetting(etcdEndpoints)
	}

	addr := ctx.String("dbaddr")
	if len(addr) <= 0 && len(config.GlobalSetting.ClickAddress) != 0 {
		addr = config.GlobalSetting.ClickAddress
	}

	if len(addr) <= 0 {
		addr = defaultClickAddr
	}
	slog.Info("Clickhouse地址：", addr)
	// local_ip
	local_ip := ctx.String("local_ip")
	if local_ip != "" {
		config.LocalIP = local_ip
	}

	// log
	log_db := ctx.String("log_db")
	log_table := ctx.String("log_table")
	ldb := logcollect.NewLogDb()
	ldb.Init(addr, config.WithDb(log_db), config.WithTable(log_table))

}

func loadConf() (ps []string) {
	ps = make([]string, 0)
	config := new(syslib.Config)
	if err := config.LoadConfig("../config.conf"); err == nil {
		apps := config.GetValueSliceString("applications")
		for _, v := range apps {
			p := path.Join("..", path.Dir(v), "logs")
			ps = append(ps, p)
		}
		return
	}
	if err := config.LoadConfig("../daemon/config.conf"); err == nil {
		apps := config.GetValueSliceString("applications")
		for _, v := range apps {
			p := path.Join(path.Dir(v), "logs")
			ps = append(ps, p)
		}
	}
	return
}
