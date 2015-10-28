package browscap_go

import (
	"bytes"
	"regexp"
	"strings"
)

var (
	rNoPrefix = regexp.MustCompile("[^a-z]")
	lenPrefix = 3
)

func inList(val []byte, list [][]byte) bool {
	for _, v := range list {
		if bytes.Equal(val, v) {
			return true
		}
	}
	return false
}

func getPrefix(s string) (prefix string) {
	if len(s) >= lenPrefix {
		prefix = s[0:lenPrefix]
	} else {
		prefix = s
	}
	prefix = strings.ToLower(prefix)
	// Fallback
	if rNoPrefix.MatchString(prefix) {
		prefix = "*"
	}
	return
}

func getNGrams(s string, n int) []string {
	if len(s) <= n {
		return []string{s}
	}

	result := make([]string, len(s)-n+1, len(s)-n+1)
	for i := 0; i <= len(s)-n; i++ {
		result[i] = s[i : i+n]
	}

	return result
}
