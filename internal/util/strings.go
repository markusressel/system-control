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

// SubstringRunes returns a substring of a string based on rune indices.
// Works like "string"[start:end], but uses rune indices instead of byte indices, which is useful for UTF-8 strings.
func SubstringRunes(s string, start int, end int) string {
	startStrIdx := 0
	i := 0
	for j := range s {
		if i == start {
			startStrIdx = j
		}
		if i == end {
			return s[startStrIdx:j]
		}
		i++
	}
	return s[startStrIdx:]
}
