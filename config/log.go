package config

import (
	"strings"

	"public/libs_go/logs"
)

func InitLog(cfg *ServerCfg) {
	appname := cfg.AppName
	ip := cfg.LocalIP
	logs.SetLogFuncCall(true)
	logs.SetLevel(cfg.LogLevel)
	logs.EnableFullFilePath(false)
	logs.WithPrefix(appname)
	logs.WithPrefix(ip)
	filepath := strings.TrimRight(cfg.LogPath, "/") + "/" + appname + ".log"
	logs.SetLogger("file", `{"filename":"`+filepath+`","daily":true,"maxdays":10,"color":false}`)
	logs.SetLogger("console")
}
