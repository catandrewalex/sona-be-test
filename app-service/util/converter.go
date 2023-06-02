package util

func BoolToInt32(b bool) int32 {
	if b {
		return 1
	} else {
		return 0
	}
}

func Int32ToBool(i int32) bool {
	return i > 0
}
