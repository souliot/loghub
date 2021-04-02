package config

import (
	"net/url"
	"strconv"

	_ "github.com/ClickHouse/clickhouse-go"
)

type ClickMode uint8

const (
	ClickStandalone ClickMode = iota
	ClickShard
	ClickShardReplica
)

type ClickSetting struct {
	Address   string
	ClickMode ClickMode
	DbName    string
	TableName string
}

type Op func(*ClickSetting)

func (c *ClickSetting) ParseAddress() {
	url, err := url.Parse(c.Address)
	if err != nil {
		return
	}
	query := url.Query()
	clickMode := query.Get("click_mode")
	ct, err := strconv.Atoi(clickMode)
	if err == nil {
		c.ClickMode = ClickMode(ct)
	}
	db := query.Get("database")
	if len(db) > 0 {
		c.DbName = db
	}
	table := query.Get("table")
	if len(table) > 0 {
		c.TableName = table
	}
}

func (c *ClickSetting) Apply(opts []Op) {
	for _, opt := range opts {
		opt(c)
	}
}

func WithDb(db string) Op {
	return func(c *ClickSetting) {
		if db != "" {
			c.DbName = db
		}
	}
}

func WithDbMode(mode ClickMode) Op {
	return func(c *ClickSetting) {
		c.ClickMode = mode
	}
}

func WithTable(t string) Op {
	return func(c *ClickSetting) {
		c.TableName = t
	}
}

type DbClickhouse struct {
	Cfg *ClickSetting
}
