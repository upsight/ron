package target

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/upsight/ron/color"
	yaml "gopkg.in/yaml.v2"
)

// Configs is a mapping of filename to target file.
type Configs struct {
	Files  []*File
	StdOut io.Writer
	StdErr io.Writer
}

// NewConfigs takes a default set of yaml in config format and then
// overrides them with a new set of config target replacements.
func NewConfigs(configs []*RawConfig, stdOut io.Writer, stdErr io.Writer) (*Configs, error) {
	if stdOut == nil {
		stdOut = os.Stdout
	}
	if stdErr == nil {
		stdErr = os.Stderr
	}

	t := &Configs{
		Files:  []*File{},
		StdOut: stdOut,
		StdErr: stdErr,
	}
	osEnvs := ParseOSEnvs(os.Environ())
	for _, config := range configs {
		var targets map[string]*Target
		if err := yaml.Unmarshal([]byte(config.Targets), &targets); err != nil {
			return nil, err
		}
		// initialize io for each target.
		for name, target := range targets {
			target.W = stdOut
			target.WErr = stdErr
			if target.Name == "" {
				target.Name = name
			}
			target.targetConfigs = t
		}

		f := &File{
			rawConfig: config,
			Filepath:  config.Filepath,
			Targets:   targets,
		}
		for _, t := range targets {
			t.File = f
		}
		e, err := NewEnv(config, osEnvs, stdOut)
		if err != nil {
			return nil, err
		}
		f.Env = e
		t.Files = append(t.Files, f)
	}
	return t, nil
}

// List prints out each target and its before and after targets.
func (tc *Configs) List(verbose bool, fuzzy string) {
	filePrefix, fuzzyTarget := splitTarget(fuzzy)
LOOP_FILES:
	for _, tf := range tc.Files {
		// If a file prefix is provided check this file matches.
		if filePrefix != "" && tf.Basename() != filePrefix {
			continue LOOP_FILES
		}

		targetNameWidth := 0
		targetNames := []string{}
	LOOP_TARGETS:
		for k := range tf.Targets {
			if len(k) > targetNameWidth {
				targetNameWidth = len(k)
			}
			if fuzzyTarget != "" {
				if ok, _ := filepath.Match(fuzzyTarget, k); !ok {
					continue LOOP_TARGETS
				}
			}
			targetNames = append(targetNames, k)
		}
		sort.Strings(targetNames)
		basename := tf.Basename()
		tc.StdOut.Write([]byte(color.Green(fmt.Sprintf("(%s) %s\n", basename, tf.Filepath))))
		for _, targetName := range targetNames {
			if target, ok := tc.Target(basename + ":" + targetName); ok {
				target.List(verbose, targetNameWidth)
			}
		}
		tc.StdOut.Write([]byte(color.Green("---\n\n")))
	}
}

// ListClean will print out a full list of targets suitable for bash completion.
func (tc *Configs) ListClean() {
	targets := []string{}
	for _, tf := range tc.Files {
		basename := tf.Basename()
		for k := range tf.Targets {
			targets = append(targets, basename+":"+k)
		}
	}
	sort.Strings(targets)
	for _, t := range targets {
		tc.StdOut.Write([]byte(t + " "))
	}
}

// Target retrieves the named target from config. If it doesn't
// exist a bool false will be returned along with nil. If the name
// contains a file prefix such as "default:mytarget", it will only
// search within that configuration file.
func (tc *Configs) Target(name string) (*Target, bool) {
	filePrefix, target := splitTarget(name)
	for _, tf := range tc.Files {
		if filePrefix != "" && tf.Basename() != filePrefix {
			continue
		}
		target, ok := tf.Targets[target]
		if ok {
			return target, ok
		}
	}
	return nil, false
}

// GetEnv will return the targets associated environment variables to
// use when running the target.
func (tc *Configs) GetEnv(name string) MSS {
	filePrefix, _ := splitTarget(name)
	for _, tf := range tc.Files {
		if filePrefix != "" && tf.Basename() != filePrefix {
			continue
		}
		envs, _ := tf.Env.Config()
		return envs
	}

	return nil
}

// ListEnvs will print out the list of file envs.
func (tc *Configs) ListEnvs() error {
	for _, tf := range tc.Files {
		tc.StdOut.Write([]byte(color.Green(fmt.Sprintf("(%s) %s\n", tf.Basename(), tf.Filepath))))
		tf.Env.List()
		tc.StdOut.Write([]byte(color.Green("---\n\n")))
	}

	return nil
}
