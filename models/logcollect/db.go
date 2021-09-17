package logcollect

import (
	"fmt"
	"net/url"
	"strings"

	"public/libs_go/ormlib/orm"

	"public/libs_go/logs"
)

var (
	clickdb           = "default"
	defaultClickhouse = new(DbClickhouse)
	DefaultUsername   = "default"
	insert_log_table  = "log.darwin_log"
	dist_log_table    = "log.darwin_log"
	DefaultLogDb      = &LogDb{
		DbName:    "log",
		TableName: "darwin_log",
	}
	defaultClickSetting = &ClickSetting{
		ClickMode: ClickStandalone,
		DbName:    "log",
		TableName: "darwin_log",
	}
)

type LogDb struct {
	DbName    string
	TableName string
}

func (m *LogDb) Init(addr string, ops ...Op) {
	cfg := defaultClickSetting
	cfg.Apply(ops)
	cfg.Address = addr
	cfg.ParseAddress()

	defaultClickhouse.Cfg = cfg
	dist_log_table = fmt.Sprintf("%s.%s", cfg.DbName, cfg.TableName)
	insert_log_table = dist_log_table

	if orm.HasDefaultDataBase() {
		clickdb = "LOG"
	}

	orm.RegisterDriver("clickhouse", orm.DRClickHouse)
	err := orm.RegisterDataBase(clickdb, "clickhouse", addr+"&read_timeout=10&write_timeout=20", true)
	if err != nil {
		logs.Error("初始化Clickhouse错误：", err)
		return
	}
	cm := cfg.ClickMode
	if cm == ClickShardReplica {
		cm = ClickShard
	}
	switch cm {
	case ClickStandalone:
		orm.RegisterModel(new(Log))
		initDBStandalone()
	case ClickShard:
		insert_log_table = fmt.Sprintf("%s_local", dist_log_table)
		orm.RegisterModel(new(Log))
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
			initDBCluster()
		}
	case ClickShardReplica:
		insert_log_table = fmt.Sprintf("%s_local", dist_log_table)
		orm.RegisterModel(new(Log))
		initDBClusterReplica()
	}

	dbs := strings.Split(insert_log_table, ".")
	if len(dbs) > 1 {
		err = orm.RegisterDataBase(clickdb, "clickhouse", addr+"&read_timeout=10&write_timeout=20&database="+dbs[0], true)
		if err != nil {
			logs.Error("初始化Clickhouse错误：", err)
			return
		}
	}
}

func initDBStandalone() {
	o := orm.NewOrm()
	o.Using(clickdb)
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + defaultClickSetting.DbName).Exec()
	if err != nil {
		logs.Error("create db <"+defaultClickSetting.DbName+"> error:", err)
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_log_table + ` (
		DateTime DateTime,
		Date Date DEFAULT toDate(DateTime),
		Level FixedString(3),
		FuncCall String,
		ServiceName String,
		Address String,
		Message String
	) ENGINE = MergeTree() PARTITION BY toYYYYMM(Date) PRIMARY KEY (DateTime)
	ORDER BY
		(DateTime) SETTINGS index_granularity = 8192
	`).Exec()
	if err != nil {
		logs.Error("create table <"+dist_log_table+"> error:", err)
		return
	}
}

func initDBCluster() {
	o := orm.NewOrm()
	o.Using("database")
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + defaultClickSetting.DbName).Exec()
	if err != nil {
		logs.Error("create db <"+defaultClickSetting.DbName+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + insert_log_table + ` (
		DateTime DateTime,
		Date Date DEFAULT toDate(DateTime),
		Level FixedString(3),
		FuncCall String,
		ServiceName String,
		Address String,
		Message String
	) ENGINE = MergeTree() PARTITION BY toYYYYMM(Date) PRIMARY KEY (DateTime)
	ORDER BY
		(DateTime) SETTINGS index_granularity = 8192
	`).Exec()
	if err != nil {
		logs.Error("create table <"+insert_log_table+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_log_table + ` AS ` + insert_log_table + ` 
	ENGINE = Distributed(cluster, ` + defaultClickSetting.DbName + `, ` + fmt.Sprintf("%s_local", defaultClickSetting.TableName) + `, rand())
	`).Exec()
	if err != nil {
		logs.Error("create table <"+dist_log_table+"> error:", err)
		return
	}
}

func initDBClusterReplica() {
	o := orm.NewOrm()
	o.Using(clickdb)
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + defaultClickSetting.DbName + ` ON CLUSTER cluster`).Exec()
	if err != nil {
		logs.Error("create db <"+defaultClickSetting.DbName+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + insert_log_table + ` ON CLUSTER cluster (
		DateTime DateTime,
		Date Date DEFAULT toDate(DateTime),
		Level FixedString(3),
		FuncCall String,
		ServiceName String,
		Address String,
		Message String
	) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{layer}-{shard}/` + fmt.Sprintf("%s_local", defaultClickSetting.TableName) + `', '{replica}') 
	PARTITION BY toYYYYMM(Date) PRIMARY KEY (DateTime)
	ORDER BY
		(DateTime) SETTINGS index_granularity = 8192
	`).Exec()
	if err != nil {
		logs.Error("create table <"+insert_log_table+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_log_table + ` ON CLUSTER cluster AS ` + insert_log_table + ` 	
  ENGINE = Distributed(cluster, ` + defaultClickSetting.DbName + `, ` + fmt.Sprintf("%s_local", defaultClickSetting.TableName) + `, rand())
	`).Exec()
	if err != nil {
		logs.Error("create table <"+dist_log_table+"> error:", err)
		return
	}
}
