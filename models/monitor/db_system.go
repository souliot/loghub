package monitor

import (
	"fmt"
	"public/libs_go/ormlib/orm"

	slog "github.com/souliot/siot-log"
)

func initSystemDbStandalone() {
	o := orm.NewOrm()
	o.Using(clickdb)
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + monitorDb.DbName).Exec()
	if err != nil {
		slog.Error("create db <"+monitorDb.DbName+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_system_table + ` (
		DateTime DateTime,
		Date Date DEFAULT toDate(DateTime),
		NodeAddress String,
		CpuStats Nested (
			Count Int32,
			Percent Float64
		),
		SysMemStats Nested (
			All UInt64,
			Used UInt64,
			Avail UInt64,
			Free UInt64
		),
		DiskStats Nested (
			Path String,
			All UInt64,
			Used UInt64,
			Free UInt64
		)
	) ENGINE = MergeTree() PARTITION BY toYYYYMM(Date) PRIMARY KEY (DateTime)
	ORDER BY
		(DateTime) SETTINGS index_granularity = 8192
	`).Exec()
	if err != nil {
		slog.Error("create table <"+dist_system_table+"> error:", err)
		return
	}
	slog.Info("初始化数据库：", monitorDb.DbName)
}

func initSystemDbCluster() {
	o := orm.NewOrm()
	o.Using("database")
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + monitorDb.DbName).Exec()
	if err != nil {
		slog.Error("create db <"+monitorDb.DbName+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + insert_system_table + ` (
		DateTime DateTime,
		Date Date DEFAULT toDate(DateTime),
		NodeAddress String,
		CpuStats Nested (
			Count Int32,
			Percent Float64
		),
		SysMemStats Nested (
			All UInt64,
			Used UInt64,
			Avail UInt64,
			Free UInt64
		),
		DiskStats Nested (
			Path String,
			All UInt64,
			Used UInt64,
			Free UInt64
		)
	) ENGINE = MergeTree() PARTITION BY toYYYYMM(Date) PRIMARY KEY (DateTime)
	ORDER BY
		(DateTime) SETTINGS index_granularity = 8192
	`).Exec()
	if err != nil {
		slog.Error("create table <"+insert_system_table+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_system_table + ` AS ` + insert_system_table + ` 
	ENGINE = Distributed(cluster, ` + monitorDb.DbName + `, ` + fmt.Sprintf("%s_local", monitorDb.SystemTable) + `, rand())
	`).Exec()
	if err != nil {
		slog.Error("create table <"+dist_system_table+"> error:", err)
		return
	}
	slog.Info("初始化数据库：", monitorDb.DbName)
}

func initSystemDbClusterReplica() {
	o := orm.NewOrm()
	o.Using(clickdb)
	_, err := o.Raw(`CREATE DATABASE IF NOT EXISTS ` + monitorDb.DbName + ` ON CLUSTER cluster`).Exec()
	if err != nil {
		slog.Error("create db <"+monitorDb.DbName+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + insert_system_table + ` ON CLUSTER cluster(
		DateTime DateTime,
		Date Date DEFAULT toDate(DateTime),
		NodeAddress String,
		CpuStats Nested (
			Count Int32,
			Percent Float64
		),
		SysMemStats Nested (
			All UInt64,
			Used UInt64,
			Avail UInt64,
			Free UInt64
		),
		DiskStats Nested (
			Path String,
			All UInt64,
			Used UInt64,
			Free UInt64
		)
	) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{layer}-{shard}/` + fmt.Sprintf("%s_local", monitorDb.SystemTable) + `', '{replica}')  
	PARTITION BY toYYYYMM(Date) PRIMARY KEY (DateTime)
	ORDER BY (DateTime) SETTINGS index_granularity = 8192
	`).Exec()
	if err != nil {
		slog.Error("create table <"+insert_system_table+"> error:", err)
		return
	}

	_, err = o.Raw(`
	CREATE TABLE IF NOT EXISTS ` + dist_system_table + ` ON CLUSTER cluster AS ` + insert_system_table + ` 
	ENGINE = Distributed(cluster, ` + monitorDb.DbName + `, ` + fmt.Sprintf("%s_local", monitorDb.SystemTable) + `, rand())
	`).Exec()
	if err != nil {
		slog.Error("create table <"+dist_system_table+"> error:", err)
		return
	}
	slog.Info("初始化数据库：", monitorDb.DbName)
}
