package config

var (
	LocalIP = ""
)

func init() {
	LocalIP = GetIPStr()
}
