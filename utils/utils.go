package utils

func InArray(search string, array []string) bool {
	for _, v := range array {
		if v == search {
			return true
		}
	}
	return false
}
