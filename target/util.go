package target

import (
	"path/filepath"
	"strings"
)

// splitTarget will take a target name and return the targets
// prefix(filename without extension) and target name.
func splitTarget(target string) (string, string) {
	tokens := strings.SplitN(target, ":", 2)
	if len(tokens) < 2 {
		// the case where there is no prefix given, just a target name.
		return "", tokens[0]
	}
	basename := filepath.Base(tokens[0])
	prefix := strings.TrimSuffix(basename, filepath.Ext(basename))
	return prefix, tokens[1]
}

func keyIn(key string, keys []string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
