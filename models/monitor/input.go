package monitor

import (
	"sync"

	"github.com/robfig/cron/v3"
)

var (
	specCollect = "@every 3s"
	msys        = make(chan MonitorSystem, 1000)
)

func monitorSystem() {
	msys <- GetMonitorSystem().(MonitorSystem)
}

type Input struct {
	c    *cron.Cron
	spec string
	lock *sync.Mutex
	jobs map[int]interface{}
}

func NewInput(spec string) (i *Input) {
	return &Input{
		c:    c,
		spec: spec,
		lock: new(sync.Mutex),
		jobs: make(map[int]interface{}, 0),
	}
}

func (m *Input) Run() {
	go m.c.Start()
	if m.spec == "" {
		m.spec = specCollect
	}
	m.AddJob(m.spec, monitorSystem)
}

func (m *Input) Stop() {
	m.c.Stop()
}

func (m *Input) AddJob(spec string, fn func()) {
	entryID, err := c.AddFunc(spec, fn)
	if err != nil {
		return
	}
	m.lock.Lock()
	m.jobs[int(entryID)] = nil
	m.lock.Unlock()
}
