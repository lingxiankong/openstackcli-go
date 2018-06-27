package util

import (
	"sort"
)

func FindString(a string, list []string) bool {
	if a == "" || len(list) == 0 {
		return false
	}

	sort.Strings(list)
	index := sort.SearchStrings(list, a)
	if index < len(list) && list[index] == a {
		return true
	}
	return false
}
