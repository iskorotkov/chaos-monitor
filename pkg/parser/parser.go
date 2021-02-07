// Package parser handles parsing string values into sets.
package parser

import (
	"strings"
)

// AsSet parses string into a set splitting it by delimiter.
func AsSet(list string, delim string) map[string]bool {
	res := make(map[string]bool)
	if list == "" {
		return res
	}

	entries := strings.Split(list, delim)
	for _, entry := range entries {
		res[entry] = true
	}

	return res
}
