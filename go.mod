module loghub

go 1.15

require (
	github.com/ClickHouse/clickhouse-go v1.4.5
	github.com/hpcloud/tail v1.0.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil v3.21.7+incompatible
	github.com/souliot/siot-log v0.0.0-20210324090643-b8655a82c129
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/sys v0.0.0-20210816074244-15123e1e1f71
	public v1.0.0
)

replace public => ../public
