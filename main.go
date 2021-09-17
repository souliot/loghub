package main

import (
	"loghub/srv"
	"os"
	"os/signal"
	"public/libs_go/servicelib"
	"runtime"
	"syscall"
	"time"
	_ "time/tzdata"

	"public/libs_go/logs"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := servicelib.Run(srv.DefaultService, srv.DefaultConf)
	if err != nil {
		logs.Error("初始化服务失败：", err)
	}
	defer func() {
		go servicelib.Stop(srv.DefaultService)
		time.Sleep(800 * time.Millisecond)
	}()

	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	_ = <-chSig
}
