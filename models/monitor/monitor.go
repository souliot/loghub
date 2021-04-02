package monitor

import (
	"time"

	"github.com/robfig/cron/v3"
)

var (
	c = cron.New()
)

type SMonitor struct {
	input  *Input
	output *Output
}

func NewMonitor(spec string, d time.Duration, gr int) (sm *SMonitor) {
	i := NewInput(spec)
	o := NewOutput(d, gr)
	return &SMonitor{
		input:  i,
		output: o,
	}
}

func (m *SMonitor) Start() {
	if m.input != nil {
		go m.input.Run()
	}
	if m.output != nil {
		go m.output.Run()
	}
}

func (m *SMonitor) Stop() {
	if m.input != nil {
		go m.input.Stop()
	}
	if m.output != nil {
		go m.output.Stop()
	}
}
