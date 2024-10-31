package util

import "strings"

func EqualsIgnoreCase(s string, substr string) bool {
	return strings.Compare(strings.ToLower(s), strings.ToLower(substr)) == 0
}

func ContainsIgnoreCase(s string, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func CountLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else {
			break
		}
	}
	return i
}

func IsNotEmpty(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) > 0
}
