package env

import (
	"strings"
)

func List(list string) map[string]bool {
	res := make(map[string]bool)
	if list == "" {
		return res
	}

	entries := strings.Split(list, ";")
	for _, entry := range entries {
		res[entry] = true
	}

	return res
}
