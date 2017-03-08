package cmd

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/pkar/runit"
)

// Command ...
type Command struct {
	Name    string
	W       io.Writer
	WErr    io.Writer
	AppName string
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	f := flag.NewFlagSet(c.Name, flag.ExitOnError)
	f.SetOutput(c.WErr)
	f.Usage = func() {
		fmt.Fprintf(c.W, "Usage: %s %s -watch <path> -wait -restart <command> -ignore <patterns>\n", c.AppName, c.Name)
		f.PrintDefaults()
	}

	alive := f.Bool("restart", false, "Restart the command if it dies.")
	watch := f.String("watch", "", "Path to directory or file to watch.")
	wait := f.Bool("wait", false, "With watch wait for file changes before running the command.")
	ignore := f.String("ignore", `.*\.git.*,.*\.DS_Store$,.*\.pyc$`, "a comma seperated list of regex patterns to ignore *optional")
	f.Parse(args)
	if f.NArg() < 1 {
		f.Usage()
		return 1, nil
	}

	ignoreList := strings.Split(*ignore, ",")
	runner, err := runit.New(strings.Join(f.Args(), " "), *watch, ignoreList, *alive, *wait)
	if err != nil {
		return 1, err
	}
	return runner.Do()
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"cmd": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Run a command with optional restart and watch for changes to restart."
}
