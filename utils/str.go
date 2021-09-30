package utils

func StringJion(str string, dem string, strs ...string) (res string) {
	tmp := []byte(str)
	for _, v := range strs {
		tmp = append(tmp, []byte("_")...)
		tmp = append(tmp, []byte(v)...)
	}
	return string(tmp)
}
