package logcollect

import "time"

type LogCollect struct {
	input  *Input
	output *Output
}

func NewLogCollect(paths []string, d time.Duration, gr int) (lc *LogCollect) {
	i := NewInput(paths, gr)
	o := NewOutput(d, gr)
	return &LogCollect{
		input:  i,
		output: o,
	}
}

func (m *LogCollect) Start() {
	if m.input != nil {
		go m.input.Run()
	}
	if m.output != nil {
		go m.output.Run()
	}
}

func (m *LogCollect) Stop() {
	if m.input != nil {
		go m.input.Stop()
	}
	if m.output != nil {
		go m.output.Stop()
	}
}
