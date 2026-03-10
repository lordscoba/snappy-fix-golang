package utils

import "strings"

func Initials(first, last string) string {
	get := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}
		r := []rune(s)
		return strings.ToUpper(string(r[0]))
	}
	return strings.TrimSpace(get(first) + get(last))
}
