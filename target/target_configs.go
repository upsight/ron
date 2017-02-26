package target

import (
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/upsight/ron/color"
	yaml "gopkg.in/yaml.v2"
)

// TargetConfigs is a mapping of filename to target file.
type TargetConfigs struct {
	Files  map[string]*TargetFile
	Env    *Env // the environment variables used for commands.
	StdOut io.Writer
	StdErr io.Writer
}

// NewTargetConfigs takes a default set of yaml in config format and then
// overrides them with a new set of config target replacements.
func NewTargetConfigs(env *Env, configs []*Config, stdOut io.Writer, stdErr io.Writer) (*TargetConfigs, error) {
	if stdOut == nil {
		stdOut = os.Stdout
	}
	if stdErr == nil {
		stdErr = os.Stderr
	}

	t := &TargetConfigs{
		Files:  map[string]*TargetFile{},
		Env:    env,
		StdOut: stdOut,
		StdErr: stdErr,
	}
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

		t.Files[config.Filepath] = &TargetFile{
			config:  config,
			Targets: targets,
		}
	}
	return t, nil
}

// List prints out each target and its before and after targets.
func (tc *TargetConfigs) List(verbose bool, fuzzy string) {
	targetNameWidth := 0
	for _, tf := range tc.Files {
		targetNames := []string{}
		for k := range tf.Targets {
			if len(k) > targetNameWidth {
				targetNameWidth = len(k)
			}
			if fuzzy != "" {
				if ok, _ := filepath.Match(fuzzy, k); !ok {
					continue
				}
			}
			targetNames = append(targetNames, k)
		}
		sort.Strings(targetNames)
		tc.StdOut.Write([]byte(color.Green(tf.config.Filepath + "\n")))
		for _, targetName := range targetNames {
			if target, ok := tc.Target(targetName); ok {
				target.List(verbose, targetNameWidth)
			}
		}
	}
}

// Target retrieves the named target from config. If it doesn't
// exists a bool false will be returned along with nil
func (tc *TargetConfigs) Target(name string) (*Target, bool) {
	for _, tf := range tc.Files {
		// FIXME this will match the first name it finds so there will
		// be collisions of similar target names.
		target, ok := tf.Targets[name]
		if ok {
			return target, ok
		}
	}

	return nil, false
}
