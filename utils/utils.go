package utils

import "strings"

func NormalizeQuotes(s string) string {
	s = strings.Replace(s, "“", "\"", -1)
	s = strings.Replace(s, "”", "\"", -1)
	s = strings.Replace(s, "‘", "'", -1)
	s = strings.Replace(s, "’", "'", -1)
	return s
}
