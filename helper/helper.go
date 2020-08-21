package helper

import "strings"

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit])
	}
	return strings.ReplaceAll(strings.ToLower(strings.Trim(strings.Replace(text, " ", "", -1), "\t \n")), "№", "n")
}
