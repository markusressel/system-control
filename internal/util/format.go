package util

import "strings"

// ReplacePlaceholders replaces placeholders in a template string with the values from a map.
// Placeholders are defined by a percent sign followed by the key and another percent sign, e.g. %key%.
func ReplacePlaceholders(template string, placeholders map[string]string) string {
	result := template
	for key, value := range placeholders {
		result = strings.ReplaceAll(result, "%"+key+"%", value)
	}
	return result
}
