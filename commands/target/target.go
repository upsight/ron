package target

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/upsight/ron/execute"
	mke "github.com/upsight/ron/make"
)

// Command ...
type Command struct {
	W       io.Writer
	WErr    io.Writer
	AppName string
	Name    string
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	f := flag.NewFlagSet(c.Name, flag.ExitOnError)
	f.Usage = func() {
		fmt.Fprintf(c.W, "Usage: %s %s\n", c.AppName, c.Name)
		f.PrintDefaults()
	}

	var defaultYamlPath string
	f.StringVar(&defaultYamlPath, "default", "", "Path to the default yaml config file, local or http.")
	var overrideYamlPath string
	f.StringVar(&overrideYamlPath, "yaml", "", `Path to override yaml file, can be local or http.

	ron contains a default set of envs and targets that can be inspected with the
	flag options listed above. Those can also be overidden with another yaml file.
	If no -default or -yaml is provided and in the current working directory there
	exists a ron.yaml, then those will be used as the -yaml option.

	The yaml config should contain a list of "envs" and a
	hash of "targets".

	env values prefixed with a +(subject to change) will be executed and set to the os environment
	prior to target execution.

		envs:
			- APP: ron
			- UNAME: +uname | tr '[:upper:]' '[:lower:]'

	targets can contain a before/after hash which is a list of other targets to
	execute. Each target should contain a cmd which can contain any valid bash
	scripting and can use previously defined envs

		targets:
			prep:
				cmd: |
					echo prep
			install:
				before:
					- prep
				after:
					- prep
				cmd: |
					echo $APP
	`)
	var listEnvs bool
	f.BoolVar(&listEnvs, "envs", false, "List the initialized environment variables.")
	var listTargets bool
	f.BoolVar(&listTargets, "list", false, "List the available targets.")
	var listTargetsShort bool
	f.BoolVar(&listTargetsShort, "l", false, "List the available targets.")
	var listTargetsClean bool
	f.BoolVar(&listTargetsClean, "list_clean", false, "List the available targets for bash completion.")
	var verbose bool
	f.BoolVar(&verbose, "verbose", false, "When used with list be verbose.")
	var verboseShort bool
	f.BoolVar(&verboseShort, "v", false, "When used with list be verbose.")
	f.BoolVar(&execute.Debug, "debug", false, "Debug the target command being run")
	f.Parse(args)
	if len(args) == 0 {
		f.Usage()
		return 1, nil
	}

	// Load default config values for envs and targets
	var err error
	defaultEnvs := mke.DefaultEnvConfig
	defaultTargets := mke.DefaultTargetConfig
	if defaultYamlPath != "" {
		defaultEnvs, defaultTargets, err = mke.LoadConfigFile(defaultYamlPath)
		if err != nil {
			return 1, err
		}
	}
	defaultEnvs = strings.TrimSpace(defaultEnvs)
	defaultTargets = strings.TrimSpace(defaultTargets)
	// Load override config values for envs and targets
	overrideEnvs := ""
	overrideTargets := ""
	if overrideYamlPath != "" {
		overrideEnvs, overrideTargets, err = mke.LoadConfigFile(overrideYamlPath)
		if err != nil {
			return 1, err
		}
	} else {
		// check if there is a default ron.yaml file to use.
		if _, err := os.Stat("ron.yaml"); err == nil {
			overrideEnvs, overrideTargets, err = mke.LoadConfigFile("ron.yaml")
			if err != nil {
				return 1, err
			}
		}
	}
	overrideEnvs = strings.TrimSpace(overrideEnvs)
	overrideTargets = strings.TrimSpace(overrideTargets)
	// Create env
	envs, err := mke.NewEnv(defaultEnvs, overrideEnvs, mke.ParseOSEnvs(os.Environ()), c.W)
	if err != nil {
		return 1, err
	}
	if listEnvs {
		err := envs.List()
		if err != nil {
			return 1, err
		}
		return 0, nil
	}
	// Create targets
	targetConfig, err := mke.NewTargetConfig(envs, defaultTargets, overrideTargets, c.W, c.WErr)
	if err != nil {
		return 1, err
	}
	if listTargets || listTargetsShort {
		targetConfig.List(verbose || verboseShort, strings.Join(f.Args(), " "))
		return 0, nil
	}
	if listTargetsClean {
		targets := []string{}
		for k := range targetConfig.Targets {
			targets = append(targets, k)
		}
		sort.Strings(targets)
		for _, t := range targets {
			targetConfig.StdOut.Write([]byte(t + " "))
		}
		return 0, nil
	}

	// Create make runner
	m, err := mke.NewMake(envs, targetConfig)
	if err != nil {
		return 1, err
	}
	err = m.Run(f.Args()...)
	if err != nil {
		return 1, err
	}

	return 0, nil
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"t":      struct{}{},
		"target": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Execute a configured target."
}
