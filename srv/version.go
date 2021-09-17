package srv

import (
	"encoding/json"
	e "public/entities"
	"public/libs_go/darwinfslib"
	"strconv"
	"strings"

	"public/libs_go/logs"
)

var moduleVersion = new(ModuleVersion)
var darwinfsProxy = new(darwinfslib.Darwinfs)

type ModuleVersion struct {
	Version *e.Version
}

//SendVersion SendVersion
func (m *ModuleVersion) CheckVersion(packet []byte) {
	if len(packet) <= 0 {
		return
	}
	ver := new(e.Version)
	err := json.Unmarshal(packet, ver)
	if err != nil {
		logs.Error("版本升级异常", err)
		return
	}

	m.Version = ver

	if !m.IsLow(version) {
		return
	}

	versionPath := "http://" + globals.GateWay + "/v1/file/" + ver.RegionID + ver.Path
	logs.Info("发现新版本", ver.Code, versionPath)
	err = darwinfsProxy.DownloadFile(versionPath, "update.zip")
	if err == nil {
		logs.Info("准备升级,程序退出")
		close()
	} else {
		logs.Error("升级失败", err)
	}
}

func (m *ModuleVersion) IsLow(code string) (isLow bool) {
	isLow = false

	curCodes := strings.Split(code, ".")
	codes := strings.Split(m.Version.Code, ".")
	for i := 0; i < len(curCodes); i++ {
		if len(codes) > i {
			curCode, _ := strconv.Atoi(curCodes[i])
			code, _ := strconv.Atoi(codes[i])
			if curCode < code {
				isLow = true
				break
			}
		}
	}

	return
}
