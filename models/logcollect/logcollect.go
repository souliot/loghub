package logcollect

import (
	"time"

	logs "github.com/souliot/siot-log"
)

var (
	LocalIP string
)

type LogCollect struct {
	input    *Input
	output   *Output
	interval time.Duration
	paths    []string
}

func NewLogCollect(paths []string, d time.Duration, gr int, local_ip string) (lc *LogCollect) {
	LocalIP = local_ip
	i := NewInput(paths, gr)
	o := NewOutput(d, gr)
	return &LogCollect{
		input:    i,
		output:   o,
		interval: d,
		paths:    paths,
	}
}

func (m *LogCollect) Start() {
	if m.input != nil {
		go m.input.Run()
	}
	if m.output != nil {
		go m.output.Run()
	}
	logs.Info("开启日志采集，采集间隔：%v，采集目录：%v", m.interval.Seconds(), m.paths)
}

func (m *LogCollect) Stop() {
	if m.input != nil {
		go m.input.Stop()
	}
	if m.output != nil {
		go m.output.Stop()
	}
	logs.Info("停止日志采集")
}
