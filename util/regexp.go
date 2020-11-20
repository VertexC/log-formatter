package util

import (
	"fmt"
	"regexp"
)

func SubMatchMapRegex(reg string, str string) (map[string]string, error) {
	r := regexp.MustCompile(reg)
	match := r.FindStringSubmatch(str)
	groupNames := r.SubexpNames()
	if len(match) != len(groupNames) {
		return nil, fmt.Errorf("Failed to extract groups %s from %s with %s, match:%s", r.SubexpNames(), str, reg, match)
	}
	subMatchMap := map[string]string{}
	for i, name := range groupNames {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}
	return subMatchMap, nil
}
