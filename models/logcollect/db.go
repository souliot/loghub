package logcollect

import (
	"fmt"
	"loghub/models/config"
	"net/url"
	"public/libs_go/ormlib/orm"
	"strings"

	slog "github.com/souliot/siot-log"
)

var (
	clickdb           = "default"
	defaultClickhouse = new(config.DbClickhouse)
	DefaultUsername   = "default"
	insert_log_table  = "log.darwin_log"
	dist_log_table    = "log.darwin_log"
	logDb             = &LogDb{
		DbName:    "log",
		TableName: "darwin_log",
	}
	defaultClickSetting = &config.ClickSetting{
		ClickMode: config.ClickStandalone,
		DbName:    "log",
		TableName: "darwin_log",
	}
)

type LogDb struct {
	DbName    string
	TableName string
}

func NewLogDb() (m *LogDb) {
	return logDb
}

func (m *LogDb) Init(addr string, ops ...config.Op) {
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
		slog.Error("初始化Clickhouse错误：", err)
		return
	}
	cm := cfg.ClickMode
	if cm == config.ClickShardReplica {
		cm = config.ClickShard
	}
	switch cm {
	case config.ClickStandalone:
		orm.RegisterModel(new(Log))
		initDBStandalone()
	case config.ClickShard:
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
	case config.ClickShardReplica:
		insert_log_table = fmt.Sprintf("%s_local", dist_log_table)
		orm.RegisterModel(new(Log))
		initDBClusterReplica()
	}

	dbs := strings.Split(insert_log_table, ".")
	if len(dbs) > 1 {
		err = orm.RegisterDataBase(clickdb, "clickhouse", addr+"&read_timeout=10&write_timeout=20&database="+dbs[0], true)
		if err != nil {
			slog.Error("初始化Clickhouse错误：", err)
			return
		}
	}
}

func initDBStandalone() {
	o := orm.NewOrm()
	o.Using(clickdb)
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + defaultClickSetting.DbName).Exec()
	if err != nil {
		slog.Error("create db <"+defaultClickSetting.DbName+"> error:", err)
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
		slog.Error("create table <"+dist_log_table+"> error:", err)
		return
	}
	slog.Info("初始化数据库：", defaultClickSetting.DbName)
}

func initDBCluster() {
	o := orm.NewOrm()
	o.Using("database")
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + defaultClickSetting.DbName).Exec()
	if err != nil {
		slog.Error("create db <"+defaultClickSetting.DbName+"> error:", err)
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
		slog.Error("create table <"+insert_log_table+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_log_table + ` AS ` + insert_log_table + ` 
	ENGINE = Distributed(cluster, ` + defaultClickSetting.DbName + `, ` + fmt.Sprintf("%s_local", defaultClickSetting.TableName) + `, rand())
	`).Exec()
	if err != nil {
		slog.Error("create table <"+dist_log_table+"> error:", err)
		return
	}
	slog.Info("初始化数据库：", defaultClickSetting.DbName)
}

func initDBClusterReplica() {
	o := orm.NewOrm()
	o.Using(clickdb)
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + defaultClickSetting.DbName + ` ON CLUSTER cluster`).Exec()
	if err != nil {
		slog.Error("create db <"+defaultClickSetting.DbName+"> error:", err)
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
		slog.Error("create table <"+insert_log_table+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_log_table + ` ON CLUSTER cluster AS ` + insert_log_table + ` 	
  ENGINE = Distributed(cluster, ` + defaultClickSetting.DbName + `, ` + fmt.Sprintf("%s_local", defaultClickSetting.TableName) + `, rand())
	`).Exec()
	if err != nil {
		slog.Error("create table <"+dist_log_table+"> error:", err)
		return
	}
	slog.Info("初始化数据库：", defaultClickSetting.DbName)
}
