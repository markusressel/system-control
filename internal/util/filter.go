package util

import (
	"path/filepath"
	"strings"
)

// GlobMatch returns true if text matches the glob pattern (case-insensitive).
// If pattern is empty, it returns true.
func GlobMatch(pattern, text string) bool {
	if pattern == "" {
		return true
	}
	matched, err := filepath.Match(strings.ToLower(pattern), strings.ToLower(text))
	if err != nil {
		return false
	}
	return matched
}

// DeviceFilter holds filter patterns for common device fields.
type DeviceFilter struct {
	Path         string
	Type         string
	Manufacturer string
	Model        string
	Serial       string
}

// Matches returns true if the given fields match the filter criteria (case-insensitive glob).
func (f DeviceFilter) Matches(path, devType, manufacturer, model, serial string) bool {
	return GlobMatch(f.Path, path) &&
		GlobMatch(f.Type, devType) &&
		GlobMatch(f.Manufacturer, manufacturer) &&
		GlobMatch(f.Model, model) &&
		GlobMatch(f.Serial, serial)
}
