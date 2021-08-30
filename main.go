package main

import (
	"loghub/config"
	"loghub/internal"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	_ "time/tzdata"

	logs "github.com/souliot/siot-log"
)

var (
	appName = "loghub"
	version = "5.2.1.0"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	srv, err := internal.NewServer(config.WithAppName(appName), config.WithVersion(version))
	if err != nil {
		logs.Error("服务启动错误：", err)
		return
	}
	srv.Start()
	defer func() {
		srv.Stop()
	}()

	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	_ = <-chSig
}
