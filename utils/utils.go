package utils

func InArray(needle string, array []string) bool {
	for _, v := range array {
		if v == needle {
			return true
		}
	}
	return false
}
