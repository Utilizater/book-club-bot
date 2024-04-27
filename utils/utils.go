package utils

import (
	"regexp"
	"strings"
)

func NormalizeQuotes(s string) string {
	s = strings.Replace(s, "“", "\"", -1)
	s = strings.Replace(s, "”", "\"", -1)
	s = strings.Replace(s, "‘", "'", -1)
	s = strings.Replace(s, "’", "'", -1)
	return s
}

func IsValidTelegramNickname(nickname string) bool {
	regex := `^[a-zA-Z][a-zA-Z0-9_]{4,31}$`
	validNickname := regexp.MustCompile(regex)
	return validNickname.MatchString(nickname)
}
