package util

import "strings"

func ReplacePlaceholders(template string, placeholders map[string]string) string {
	result := template
	for key, value := range placeholders {
		result = strings.ReplaceAll(result, "%"+key+"%", value)
	}
	return result
}
