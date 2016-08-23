package ron

import (
	"fmt"
	"io"
)

// Command ...
type Command struct {
	W          io.Writer
	WErr       io.Writer
	AppName    string
	AppVersion string
	GitCommit  string
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	fmt.Fprintln(c.W, c.AppName, c.AppVersion, c.GitCommit)
	return 0, nil
}

// Names are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Names() map[string]struct{} {
	return map[string]struct{}{
		"version": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Print the version."
}
