package controllers

import (
	"loghub/models"
	"loghub/models/config"
	"loghub/models/logcollect"
	"loghub/models/monitor"
	"time"

	"github.com/urfave/cli/v2"
)

var LogHubController = new(LogHub)
var lc *logcollect.LogCollect
var mt *monitor.SMonitor

type LogHub struct{}

func (c *LogHub) Run(ctx *cli.Context) {
	c.InitBase(ctx)
	log_paths := ctx.StringSlice("log_paths")
	log_interval := ctx.Int64("log_interval")
	monitor_spec := ctx.String("monitor_spec")
	lc = logcollect.NewLogCollect(log_paths, time.Duration(log_interval)*time.Second, models.GoPoolSize)
	lc.Start()

	mt = monitor.NewMonitor(monitor_spec, time.Duration(log_interval)*time.Second, models.GoPoolSize)
	mt.Start()
}

func (c *LogHub) Stop() {
	lc.Stop()
	mt.Stop()
}

func (c *LogHub) InitBase(ctx *cli.Context) {
	// 设置goroutine size
	models.GoPoolSize = ctx.Int("gopoolsize")
	models.TestMode = ctx.Bool("test")
	addr := ctx.String("dbaddr")

	local_ip := ctx.String("local_ip")
	if local_ip != "" {
		config.LocalIP = local_ip
	}

	// log
	log_db := ctx.String("log_db")
	log_table := ctx.String("log_table")
	ldb := logcollect.NewLogDb()
	ldb.Init(addr, config.WithDb(log_db), config.WithTable(log_table))

	// monitor
	db := ctx.String("monitor_db")
	sys := ctx.String("system_table")
	mdb := monitor.NewMonitorDb(db, sys)
	mdb.Init(addr)
}
