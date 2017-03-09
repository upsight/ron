package target

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

	switch {
	case runtime.GOOS == "windows":
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	case runtime.GOOS == "linux":
		var stdout bytes.Buffer
		cmd := exec.Command("getent", "passwd", strconv.Itoa(os.Getuid()))
		cmd.Stdout = &stdout
		if err := cmd.Run(); err != nil {
			return ""
		}
		passwd := strings.TrimSpace(stdout.String())
		if passwd == "" {
			return ""
		}
		// travis:x:1000:1000::/home/travis:/bin/bash
		tokens := strings.SplitN(passwd, ":", 7)
		if len(tokens) > 5 {
			return tokens[5]
		}
		return ""
	}

	// NOTE: `cd && pwd` will fail if $HOME is not set so this will never
	// work. Here for reference
	// sh: line 0: cd: HOME not set
	//
	// stdout.Reset()
	// var stderr bytes.Buffer
	// cmd := exec.Command("sh", "-c", "cd && pwd")
	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr
	// if err := cmd.Run(); err != nil {
	// 	fmt.Println(stderr.String())
	// 	return ""
	// }

	return ""
}
