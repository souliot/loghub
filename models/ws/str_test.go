package ws

import (
	"fmt"
	"testing"
)

func BenchmarkStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringJion("llz", "_", "asdasda", "qweaczxc", "asdasdqw")
	}
}
func BenchmarkStr2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		str2()
	}
}

func StringJion(str string, dem string, strs ...string) (res string) {
	tmp := []byte(str)
	for _, v := range strs {
		tmp = append(tmp, []byte("_")...)
		tmp = append(tmp, []byte(v)...)
	}
	return string(tmp)
}

func str2() (res string) {
	id := "llz"
	id2 := "asdasd"
	id3 := "qwetercv"
	return fmt.Sprintf("%s%s%s", id, id2, id3)
}
