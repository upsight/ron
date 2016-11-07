package ron

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/upsight/ron/commands/bash"
	"github.com/upsight/ron/commands/burgundy"
	"github.com/upsight/ron/commands/cmd"
	hs "github.com/upsight/ron/commands/httpstat"
	rpl "github.com/upsight/ron/commands/replace"
	trg "github.com/upsight/ron/commands/target"
	tmpl "github.com/upsight/ron/commands/template"
	upg "github.com/upsight/ron/commands/upgrade"
	vrs "github.com/upsight/ron/commands/version"
)

// Command defines the methods a command should implement.
type Command interface {
	Run([]string) (int, error)
	Names() map[string]struct{}
	Description() string
}

// Commander is a mapping of name to command.
type Commander []Command

// NewCommander creates a new map of command objects, where the keys are the
// names of the commands.
func NewCommander(stdOut io.Writer, stdErr io.Writer) Commander {
	if stdOut == nil {
		stdOut = os.Stdout
	}
	if stdErr == nil {
		stdErr = os.Stderr
	}

	return Commander{
		&bash.Command{W: stdOut, WErr: stdErr},
		&burgundy.Command{W: stdOut, WErr: stdErr},
		&cmd.Command{W: stdOut, WErr: stdErr, AppName: AppName, Name: "cmd"},
		&hs.Command{W: stdOut, WErr: stdErr, AppName: AppName, Name: "httpstat"},
		&rpl.Command{W: stdOut, WErr: stdErr, AppName: AppName, Name: "replace"},
		&trg.Command{W: stdOut, WErr: stdErr, AppName: AppName, Name: "target"},
		&tmpl.Command{W: stdOut, WErr: stdErr, AppName: AppName, Name: "template"},
		&upg.Command{W: stdOut, WErr: stdErr, AppName: AppName},
		&vrs.Command{W: stdOut, WErr: stdErr, AppName: AppName, AppVersion: AppVersion, GitCommit: GitCommit},
	}
}

type nameDescription struct {
	name        string
	description string
}

// Usage displays the available commands.
func (c Commander) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "Usage: %s <command>\n\nAvailable commands are:\n", AppName)
	maxLength := 0
	metas := []*nameDescription{}
	for _, cmd := range c {
		names := []string{}
		for k := range cmd.Names() {
			names = append(names, k)
		}
		sort.Strings(names)
		name := strings.Join(names, ", ")
		if len(name) > maxLength {
			maxLength = len(name)
		}
		metas = append(metas, &nameDescription{name, cmd.Description()})
	}
	format := fmt.Sprintf("    %%-%ds    %%s\n", maxLength)
	for _, m := range metas {
		fmt.Fprintf(writer, format, m.name, m.description)
	}
	fmt.Fprintf(writer, "\n")
}
