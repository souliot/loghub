package monitor

import (
	"fmt"
	"loghub/models/config"
	"net/url"
	"strings"

	"public/libs_go/ormlib/orm"

	slog "github.com/souliot/siot-log"
)

var (
	clickdb         = "default"
	DefaultUsername = "default"

	insert_system_table = "monitor.system"
	dist_system_table   = "monitor.system"

	monitorDb = &MonitorDb{
		DbName:      "monitor",
		SystemTable: "system",
	}

	defaultClickSetting = &config.ClickSetting{
		ClickMode: config.ClickStandalone,
		DbName:    "log",
		TableName: "darwin_log",
	}
)

type MonitorDb struct {
	DbName      string
	SystemTable string
}

func NewMonitorDb(db, sys string) (m *MonitorDb) {
	monitorDb.DbName = db
	monitorDb.SystemTable = sys

	return monitorDb
}

func (m *MonitorDb) Init(addr string, ops ...config.Op) {
	cfg := defaultClickSetting
	cfg.Apply(ops)
	cfg.Address = addr
	cfg.ParseAddress()

	dist_system_table = fmt.Sprintf("%s.%s", m.DbName, m.SystemTable)
	insert_system_table = dist_system_table

	if orm.HasDefaultDataBase() {
		clickdb = "MONITOR"
	}
	// orm.ReleaseDataBase("default")
	orm.RegisterDriver("clickhouse", orm.DRClickHouse)
	err := orm.RegisterDataBase(clickdb, "clickhouse", addr+"&read_timeout=10&write_timeout=20", true)
	if err != nil {
		slog.Error("初始化Clickhouse错误：", err)
		return
	}
	cm := cfg.ClickMode
	if cm == config.ClickShardReplica {
		cm = config.ClickShard
	}
	switch cm {
	case config.ClickStandalone:
		initSystemDbStandalone()
	case config.ClickShard:
		insert_system_table = fmt.Sprintf("%s_local", dist_system_table)
		url, err := url.Parse(cfg.Address)
		if err != nil {
			return
		}
		hosts := []string{url.Host}
		query := url.Query()
		username := query.Get("username")
		if len(username) == 0 {
			username = DefaultUsername
		}
		password := query.Get("password")
		if altHosts := strings.Split(query.Get("alt_hosts"), ","); len(altHosts) != 0 {
			for _, host := range altHosts {
				if len(host) != 0 {
					hosts = append(hosts, host)
				}
			}
		}
		for _, v := range hosts {
			orm.RegisterDataBase("database", "clickhouse", "tcp://"+v+"?username="+username+"&password="+password+"&read_timeout=10&write_timeout=20", true)
			initSystemDbCluster()
		}
	case config.ClickShardReplica:
		insert_system_table = fmt.Sprintf("%s_local", dist_system_table)
		initSystemDbClusterReplica()
	}
}
