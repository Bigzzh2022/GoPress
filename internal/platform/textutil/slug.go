package textutil

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func Slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	builder.Grow(len(value))
	separator := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if separator && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			separator = false
			builder.WriteRune(r)
		case r == '-' || r == '_' || unicode.IsSpace(r) || unicode.IsPunct(r):
			separator = true
		default:
			separator = true
		}
	}
	return truncateRunes(builder.String(), 180)
}

func truncateRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return strings.TrimRight(string(runes[:limit]), "-")
}
