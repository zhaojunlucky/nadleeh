package util

func Bool2Int(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}

func Str2Bool(val string) bool {
	return val == "true"
}

func Int2Bool(val int64) bool {
	return val >= 1
}
