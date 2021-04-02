package monitor

import (
	"loghub/models/config"
	"public/libs_go/ormlib/orm"
	"sync"
	"sync/atomic"
	"time"

	slog "github.com/souliot/siot-log"
)

var (
	db_chs       chan bool
	nodeAddress  string
	monitor_bulk int64 = 5000
)

type Output struct {
	gr     int
	ticker *time.Ticker
	lock   *sync.Mutex
}

func NewOutput(d time.Duration, gr int) (o *Output) {
	return &Output{
		gr:     gr,
		ticker: time.NewTicker(d),
		lock:   new(sync.Mutex),
	}
}

func (m *Output) Run() {
	db_chs = make(chan bool, m.gr)
	o_sys := make([]MonitorSystem, 0)
	nodeAddress = config.LocalIP
	var cnt int64 = 0

	go func() {
		for {
			select {
			case <-m.ticker.C:
				datas := make([]MonitorSystem, len(o_sys))
				copy(datas, o_sys)
				db_chs <- true
				go insertSystemMonitor(datas)
				m.lock.Lock()
				o_sys = make([]MonitorSystem, 0)
				m.lock.Unlock()
				atomic.StoreInt64(&cnt, 0)
			}
		}
	}()
	for sys := range msys {
		o_sys = append(o_sys, sys)
		atomic.AddInt64(&cnt, 1)
		if cnt >= monitor_bulk {
			datas := make([]MonitorSystem, len(o_sys))
			copy(datas, o_sys)
			db_chs <- true
			go insertSystemMonitor(datas)
			m.lock.Lock()
			o_sys = make([]MonitorSystem, 0)
			m.lock.Unlock()
			atomic.StoreInt64(&cnt, 0)
		}
	}
}

func (m *Output) Stop() {
	m.ticker.Stop()
}

func insertSystemMonitor(msys []MonitorSystem) (cnt int64, err error) {
	defer func() {
		<-db_chs
	}()

	if len(msys) <= 0 {
		return
	}
	cnt = 0
	o := orm.NewOrm()
	o.Using(clickdb)
	o.Begin()

	p, err := o.Raw(`INSERT INTO ` + insert_system_table + `
		(DateTime, NodeAddress,
		CpuStats.Count, CpuStats.Percent,
		SysMemStats.All, SysMemStats.Used, SysMemStats.Avail, SysMemStats.Free,  
		DiskStats.Path, DiskStats.All, DiskStats.Used, DiskStats.Free)
		VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).Prepare()
	if err != nil {
		slog.Error(err)
		return
	}
	for _, v := range msys {
		dpath := make([]string, 0)
		dall := make([]uint64, 0)
		dused := make([]uint64, 0)
		dfree := make([]uint64, 0)
		for _, d := range v.DiskStats {
			dpath = append(dpath, d.Path)
			dall = append(dall, d.All)
			dused = append(dused, d.Used)
			dfree = append(dfree, d.Free)
		}
		if _, err = p.Exec(
			v.DateTime,
			nodeAddress,
			[]int{v.CpuStats.Count},
			[]float64{v.CpuStats.Percent},
			[]uint64{v.SysMemStats.All},
			[]uint64{v.SysMemStats.Used},
			[]uint64{v.SysMemStats.Avail},
			[]uint64{v.SysMemStats.Free},
			dpath,
			dall,
			dused,
			dfree,
		); err == nil {
			cnt += 1
		} else {
			slog.Error(err)
		}
	}

	o.Commit()
	defer p.Close()
	return
}

// func insertServiceMonitor() (cnt int64, err error) {
// 	o := orm.NewOrm()
// 	o.Using(clickdb)
// 	o.Begin()

// 	p, err := o.Raw(`INSERT INTO ` + insert_service_table + `
// 		(DateTime, NodeAddress, ServiceType, Cmdline, RunTime, GoVersion, CPUs, OS, Goroutines, Cgos,
// 		Memstats.Sys, Memstats.Alloc, Memstats.HeapSys, Memstats.HeapAlloc, Memstats.HeapIdle, Memstats.HeapInuse, Memstats.HeapReleased, Memstats.StackSys, Memstats.StackInuse, Memstats.GCSys, Memstats.OtherSys, Memstats.NextGC)
// 		VALUES
// 		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).Prepare()
// 	if err != nil {
// 		return
// 	}
// 	for _, v := range msrv {
// 		if _, err = p.Exec(
// 			v.DateTime,
// 			serviceAddress,
// 			serviceType,
// 			v.Cmdline,
// 			v.RunTime,
// 			v.GoVersion,
// 			v.CPUs, v.OS,
// 			v.Goroutines,
// 			v.Cgos,
// 			[]uint64{v.Memstats.Sys},
// 			[]uint64{v.Memstats.Alloc},
// 			[]uint64{v.Memstats.HeapSys},
// 			[]uint64{v.Memstats.HeapAlloc},
// 			[]uint64{v.Memstats.HeapIdle},
// 			[]uint64{v.Memstats.HeapInuse},
// 			[]uint64{v.Memstats.HeapReleased},
// 			[]uint64{v.Memstats.StackSys},
// 			[]uint64{v.Memstats.StackInuse},
// 			[]uint64{v.Memstats.GCSys},
// 			[]uint64{v.Memstats.OtherSys},
// 			[]uint64{v.Memstats.NextGC},
// 		); err == nil {
// 			cnt += 1
// 		} else {
// 			slog.Error(err)
// 		}
// 	}

// 	o.Commit()
// 	defer p.Close()

// 	return
// }
