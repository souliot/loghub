package main

import (
	"loghub/controllers"
	"loghub/models/config"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	_ "time/tzdata"

	slog "github.com/souliot/siot-log"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

var (
	appName  = "loghub"
	describe = "a loghub for format log file"
	version  = "5.1.2.0"
	logpath  = "logs"
)

func init() {
	slog.SetLogFuncCall(true)
	slog.EnableFullFilePath(false)
	slog.SetLogFuncCallDepth(3)
	slog.SetLevel(slog.LevelInfo)
	slog.Async()
	slog.WithPrefix("loghub")
	slog.WithPrefix(config.GetIPStr())
	filepath := strings.TrimRight(logpath, "/") + "/loghub.log"
	slog.SetLogger("file", `{"filename":"`+filepath+`","daily":true,"maxdays":10,"color":false}`)
	slog.SetLogger("console")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	start()
	defer func() {
		stop()
		time.Sleep(200 * time.Millisecond)
	}()
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	_ = <-chSig
}

func start() {
	slog.Info("服务版本号：", version)
	app := cli.NewApp()
	app.Name = appName
	app.Usage = describe
	app.Version = version
	app.Commands = []*cli.Command{}
	app.Flags = []cli.Flag{
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:    "test",
			Aliases: []string{"t"},
			Value:   false,
			Usage:   "Test for log ``",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "local_ip",
			Aliases: []string{"ip"},
			Value:   "",
			Usage:   "Used Local IP to monitor `local_ip`",
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name:    "etcdendpoints",
			Aliases: []string{"ep"},
			Value:   cli.NewStringSlice("127.0.0.1:2379"),
			Usage:   "Used Etcd Endpoints `etcdendpoints`",
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "gopoolsize",
			Aliases: []string{"g"},
			Value:   runtime.NumCPU(),
			Usage:   "Goroutine size of the program `Size`",
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "log_interval",
			Aliases: []string{"li"},
			Value:   10,
			Usage:   "Log insert Interval `Interval`",
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name:    "log_paths",
			Aliases: []string{"lp"},
			Value:   cli.NewStringSlice("logs"),
			Usage:   "Used log paths `log_paths`",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "dbaddr",
			Aliases: []string{"a"},
			Value:   "",
			Usage:   "Used clickhouse `dbaddress`",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "log_db",
			Aliases: []string{"ldb"},
			Value:   "log",
			Usage:   "Used clickhouse `log_db`",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "log_table",
			Aliases: []string{"ltb"},
			Value:   "darwin_log",
			Usage:   "Used clickhouse `log_table`",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "monitor_db",
			Aliases: []string{"mdb"},
			Value:   "monitor",
			Usage:   "Used clickhouse `monitor_db`",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "monitor_spec",
			Aliases: []string{"ms"},
			Value:   "@every 5s",
			Usage:   "Used monitor spec `monitor_spec`",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "system_table",
			Aliases: []string{"systb"},
			Value:   "system",
			Usage:   "Used clickhouse `system_table`",
		}),
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   "config.yaml",
			Usage:   "Load the config file `YAML`",
		},
	}
	app.Action = func(c *cli.Context) (err error) {
		controllers.LogHubController.Run(c)
		return
	}
	app.Before = altsrc.InitInputSourceWithContext(app.Flags, controllers.NewYamlSourceFromFlagFunc("config"))
	err := app.Run(os.Args)
	if err != nil {
		slog.Error(err)
	}
}

func stop() {
	controllers.LogHubController.Stop()
}
