package logic

import (
	"regexp"
	"strings"
)

var badFileChars = regexp.MustCompile(`[\\/:*?"<>|]+`)

func sanitizeFilePart(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ",", "")
	s = badFileChars.ReplaceAllString(s, "_")
	return s
}
