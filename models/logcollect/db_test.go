package logcollect

import (
	"loghub/models/config"
	"testing"
)

var (
	// addr = "tcp://192.168.0.4:9008?username=default&password=watrix888"
	addr = "tcp://192.168.2.36:9008?alt_hosts=192.168.2.37:9008,192.168.2.38:9008&username=default&password=watrix888&click_mode=2"
)

func TestInitLog(t *testing.T) {
	InitLogDb(addr, config.WithDb("log_test"), config.WithTable("darwin_log"))
}
