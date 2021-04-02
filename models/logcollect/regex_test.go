package logcollect

import (
	"testing"
)

func TestRegex(t *testing.T) {
	lr, err := ParseLineRegex(lineRegex[2])
	if err != nil {
		t.Fatal(err)
	}
	match := lr.FindAllStringSubmatch("2021/04/01 19:19:05.674  [I]  [input.go:100] [0] [192.168.0.252]  收到日志信息...", -1)
	parsed := make(map[string]interface{})
	var firstMatch []string = match[0]
	for i, name := range lr.SubexpNames() {
		if i != 0 && i < len(firstMatch) {
			parsed[name] = firstMatch[i]
		}
	}
	t.Log(parsed)
}

func TestRegexSelf(t *testing.T) {
	lr, err := NewRegexLineParser(lineRegex)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := lr.ParseLine("2021/04/01 19:41:01.517  [I]  [db.go:187]  [0] [192.168.0.252]  初始化数据库： log_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(parsed)
}

func TestFileName(t *testing.T) {
	file := "business/logs/business.info.2010.log"

	t.Log(getFileName(file))
}
