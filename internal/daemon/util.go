package daemon

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateStatePath automatically generates a standardized tmpfs file path
// based on the command ID and its active flags.
func GenerateStatePath(commandId string, flags map[string]string) string {
	path := strings.ReplaceAll(commandId, ".", "/")

	if len(flags) > 0 {
		var keys []string
		for k := range flags {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var flagParts []string
		for _, k := range keys {
			v := flags[k]
			if v != "" {
				flagParts = append(flagParts, fmt.Sprintf("%s-%s", k, v))
			}
		}

		if len(flagParts) > 0 {
			path += "_" + strings.Join(flagParts, "_")
		}
	}

	return path
}
