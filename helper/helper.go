package helper

import "strings"

// ContainsString склейка массивов
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Cut обрезка текста
func Cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		runes = runes[:limit]
	}
	return strings.ReplaceAll(strings.ToLower(strings.Trim(strings.Replace(string(runes), " ", "", -1), "\t \n")), "№", "n")
}
