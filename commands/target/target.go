package target

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/upsight/ron/execute"
	"github.com/upsight/ron/target"
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
	If no -default or -yaml is provided and in the current or parent working directory there
	exists a ron.yaml, then those will be used as the -yaml option.

	The yaml config should contain "remotes" (optional), "envs", and a hash of "targets".

	remotes should be defined as a map with any environment name and a list of server values. It's only
	necessary to define them once so they could be globally set for example in ~/.ron/remotes.yaml
	You can then reference it with -remote=remotes:some_other_env

		remotes:
			staging:
				-
					host: example1.com
					port: 22
					user: test
				-
					host: example2.com
					port: 22
					user: test
			some_other_env:
				-
					host: exampleprod.com
					port: 22
					user: test
					proxy_host: bastionserver.com
					proxy_port: 22
					proxy_user: bastion_user
					identity_file: /optional/path/to/identityfile

	If no identity file is provided, the users local ssh agent will be attempted. You can add
	keys with ssh-add.

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
	var listRemotes bool
	f.BoolVar(&listRemotes, "list_remotes", false, "List the initialized remotes configurations.")
	var listTargets bool
	f.BoolVar(&listTargets, "list", false, "List the available targets.")
	var listTargetsShort bool
	f.BoolVar(&listTargetsShort, "l", false, "List the available targets.")
	var listTargetsClean bool
	f.BoolVar(&listTargetsClean, "list_clean", false, "List the available targets for bash completion.")
	var remoteEnv string
	f.StringVar(&remoteEnv, "remotes", "", "The remote target environment to run the target on.")
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

	// FIXME When using globs and os.Args without quoting the target, the glob will match
	// any files in the directory.
	//
	// go run cmd/ron/main.go t -default=target/default.yaml -l b*
	// [/var/folders/rs/0jn_2dpn36x53x8tvgvptr740000gn/T/go-build477108253/command-line-arguments/_obj/exe/main t -default=target/default.yaml -l bin]
	// go run cmd/ron/main.go t -default=target/default.yaml -l "b*"
	// [/var/folders/rs/0jn_2dpn36x53x8tvgvptr740000gn/T/go-build906759632/command-line-arguments/_obj/exe/main t -default=target/default.yaml -l b*]

	configs, foundConfigDir, err := target.LoadConfigFiles(defaultYamlPath, overrideYamlPath, true)
	if err != nil {
		return 1, err
	}
	if foundConfigDir != "" {
		// If we discovered a config in a parent folder, change the working
		// directory to that folder so Ron targets run from the expected place.
		os.Chdir(foundConfigDir)
	}
	// Create targets
	targetConfig, err := target.NewConfigs(configs, remoteEnv, c.W, c.WErr)
	if err != nil {
		return 1, err
	}
	if listTargets || listTargetsShort {
		targetConfig.List(verbose || verboseShort, strings.Join(f.Args(), " "))
		return 0, nil
	}
	if listTargetsClean {
		targetConfig.ListClean()
		return 0, nil
	}
	if listEnvs {
		err := targetConfig.ListEnvs()
		if err != nil {
			return 1, err
		}
		return 0, nil
	}
	if listRemotes {
		err := targetConfig.ListRemotes()
		if err != nil {
			return 1, err
		}
		return 0, nil
	}

	// Create make runner
	m, err := target.NewMake(targetConfig)
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
