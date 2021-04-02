package config

import (
	"encoding/json"
	"os"
	"public/libs_go/gateway/master"
	"time"

	slog "github.com/souliot/siot-log"
)

var (
	GlobalSetting = new(Config)
)

type Config struct {
	ClickAddress string `bson:"ClickAddress"`
}

func WatchGlobalSetting(etcdEndpoints []string) {
	setting_name := "AppSetting"
	timeout := 10 * time.Second
	ms, err := master.OnWatchSetting(etcdEndpoints, setting_name, timeout)
	if err != nil {
		slog.Error("初始化服务配置失败：", err)
		os.Exit(0)
		return
	}
	err = json.Unmarshal(ms.Value, GlobalSetting)
	if err != nil {
		slog.Error("初始化服务配置失败：", err)
		os.Exit(0)
		return
	}
	go func() {
		for {
			select {
			case <-ms.IsUpdate:
				GlobalSetting = new(Config)
				err = json.Unmarshal(ms.Value, GlobalSetting)
				if err != nil {
					slog.Error("解析配置失败：", err)
					continue
				} else {
					slog.Info("更新配置：" + string(ms.Value))
					os.Exit(0)
				}
			}
		}
	}()
}
