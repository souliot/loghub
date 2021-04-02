module loghub

go 1.15

require (
	github.com/ClickHouse/clickhouse-go v1.4.0
	github.com/hpcloud/tail v1.0.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil v2.20.3+incompatible
	github.com/souliot/siot-log v0.0.0-20210324090643-b8655a82c129
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/net v0.0.0-20200625001655-4c5254603344 // indirect
	golang.org/x/sys v0.0.0-20210331175145-43e1dd70ce54
	public v0.0.0-00010101000000-000000000000
)

replace public => ../public
