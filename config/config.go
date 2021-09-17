package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"

	"public/libs_go/logs"
	"github.com/spf13/viper"
)

var (
	errConfigNotInit = fmt.Errorf("config have not init")
)

type ServerCfg struct {
	Id            string     `mapstructure:"id"`
	AppName       string     `mapstructure:"appname"`
	LogLevel      int        `mapstructure:"loglevel"`
	LogPath       string     `mapstructure:"logpath"`
	LocalIP       string     `mapstructure:"localip"`
	HttpPort      int        `mapstructure:"httpport"`
	EtcdEndpoints []string   `mapstructure:"etcdendpoints"`
	GoPoolSize    int        `mapstructure:"gopoolsize"`
	Collector     *Collector `mapstructure:"collector"`
}

type Collector struct {
	Interval  int      `mapstructure:"interval"`
	Paths     []string `mapstructure:"paths"`
	DBName    string   `mapstructure:"dbname"`
	TableName string   `mapstructure:"tablename"`
}

var Config *viper.Viper

func InitConfig() (err error) {
	Config = viper.New()
	Config.SetConfigType("yaml")
	b, err := json.Marshal(DefaultServerCfg)
	if err != nil {
		return
	}
	defaultConfig := bytes.NewReader(b)
	Config.ReadConfig(defaultConfig)
	Config.SetConfigFile("config.yaml")
	err = Config.ReadInConfig()
	if err != nil {
		logs.Info("Using default config")
	} else {
		Config.MergeInConfig()
	}

	err = Config.Unmarshal(DefaultServerCfg)
	if err != nil {
		return
	}
	if DefaultServerCfg.LocalIP == "" {
		DefaultServerCfg.LocalIP = GetIPStr()
	}
	return
}

type Option func(*ServerCfg)

var DefaultServerCfg = &ServerCfg{
	AppName:       "loghub",
	LogLevel:      logs.LevelInfo,
	LogPath:       "logs",
	LocalIP:       "",
	HttpPort:      8890,
	EtcdEndpoints: []string{},
	GoPoolSize:    runtime.NumCPU(),
	Collector: &Collector{
		Interval:  10,
		Paths:     []string{"logs"},
		DBName:    "log",
		TableName: "darwin_log",
	},
}

func (c *ServerCfg) Apply(opts []Option) {
	for _, opt := range opts {
		opt(c)
	}
}

func (c *ServerCfg) SaveConfigFile() (err error) {
	if Config == nil {
		return errConfigNotInit
	}
	err = Config.WriteConfigAs(Config.ConfigFileUsed())
	return
}

func WithAppName(name string) Option {
	return func(c *ServerCfg) {
		c.AppName = name
	}
}
