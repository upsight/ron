package target

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// homeDir will attempt to find the home directory of the current
// user. An empty string returned means the users home directory
// could not be found.
func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "cd && pwd")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(stdout.String())
}
