package logcollect

import (
	"public/libs_go/ormlib/orm"
	"sync"
	"sync/atomic"
	"time"

	slog "github.com/souliot/siot-log"
)

var (
	db_chs   chan bool
	log_bulk int64 = 5000
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
	o_logs := make([]*Log, 0)
	var cnt int64 = 0
	go func() {
		for {
			select {
			case <-m.ticker.C:
				datas := make([]*Log, len(o_logs))
				copy(datas, o_logs)
				db_chs <- true
				go insertLog(datas)
				m.lock.Lock()
				o_logs = make([]*Log, 0)
				m.lock.Unlock()
				atomic.StoreInt64(&cnt, 0)
			}
		}
	}()
	for log := range mlog {
		o_logs = append(o_logs, log)
		atomic.AddInt64(&cnt, 1)
		if cnt >= log_bulk {
			datas := make([]*Log, len(o_logs))
			copy(datas, o_logs)
			db_chs <- true
			go insertLog(datas)
			m.lock.Lock()
			o_logs = make([]*Log, 0)
			m.lock.Unlock()
			atomic.StoreInt64(&cnt, 0)
		}
	}
}

func (m *Output) Stop() {
	m.ticker.Stop()
}

func insertLog(datas []*Log) (cnt int64, err error) {
	defer func() {
		<-db_chs
	}()

	if len(datas) <= 0 {
		return
	}
	cnt = 0
	o := orm.NewOrm()
	o.Using(clickdb)
	count, err := o.InsertMulti(5000, datas)
	if err != nil {
		slog.Error("写入clickhouse日志错误：", err)
		return
	}
	cnt = count.(int64)
	datas = nil
	return
}
