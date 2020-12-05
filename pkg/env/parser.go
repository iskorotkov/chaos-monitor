package env

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func ParseNames(crashTolerance string) map[string]int {
	res := make(map[string]int)

	if crashTolerance == "" {
		return res
	}

	entries := strings.Split(crashTolerance, ";")
	for _, entry := range entries {
		kv := strings.Split(entry, "=")

		if len(kv) != 2 {
			log.Fatal(fmt.Sprintf("couldn't split '%s' on key-value pair", entry))
		}

		key, value := kv[0], kv[1]

		num, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			log.Fatal(fmt.Sprintf("couldn't parse '%v' to int", num))
		}

		res[key] = int(num)
	}

	return res
}

func ParseLabels(crashTolerance string) map[string]int {
	res := make(map[string]int)

	if crashTolerance == "" {
		return res
	}

	entries := strings.Split(crashTolerance, ";")
	for _, entry := range entries {
		kv := strings.Split(entry, "=")

		if len(kv) != 3 {
			log.Fatal(fmt.Sprintf("couldn't split '%s' on 3 parts", entry))
		}

		label := fmt.Sprintf("%s=%s", kv[0], kv[1])
		value := kv[2]

		num, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			log.Fatal(fmt.Sprintf("couldn't parse '%v' to int", num))
		}

		res[label] = int(num)
	}

	return res
}
