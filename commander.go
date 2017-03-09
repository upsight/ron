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
	"github.com/upsight/ron/commands/httpstat"
	"github.com/upsight/ron/commands/replace"
	"github.com/upsight/ron/commands/target"
	tmpl "github.com/upsight/ron/commands/template"
	upg "github.com/upsight/ron/commands/upgrade"
	vrs "github.com/upsight/ron/commands/version"
)

// Command defines the methods a command should implement.
type Command interface {
	Description() string
	Key() string
	Aliases() map[string]struct{}
	Run([]string) (int, error)
}

type nameDescription struct {
	name        string
	description string
}

// Commander is a mapping of name to command.
type Commander struct {
	Stderr   io.Writer
	Stdout   io.Writer
	Commands []Command
}

// NewDefaultCommander creates a new map of command objects, where the keys are the
// names of the commands.
func NewDefaultCommander(stdOut io.Writer, stdErr io.Writer) *Commander {
	if stdOut == nil {
		stdOut = os.Stdout
	}
	if stdErr == nil {
		stdErr = os.Stderr
	}

	c := &Commander{
		Stderr: stdErr,
		Stdout: stdOut,
		Commands: []Command{
			&bash.Command{Name: "bash", W: stdOut, WErr: stdErr},
			&burgundy.Command{Name: "burgundy", W: stdOut, WErr: stdErr},
			&cmd.Command{Name: "cmd", W: stdOut, WErr: stdErr, AppName: AppName},
			&httpstat.Command{Name: "httpstat", W: stdOut, WErr: stdErr, AppName: AppName},
			&replace.Command{Name: "replace", W: stdOut, WErr: stdErr, AppName: AppName},
			&target.Command{Name: "target", W: stdOut, WErr: stdErr, AppName: AppName},
			&tmpl.Command{Name: "template", W: stdOut, WErr: stdErr, AppName: AppName},
			&upg.Command{Name: "upgrade", W: stdOut, WErr: stdErr, AppName: AppName},
			&vrs.Command{Name: "version", W: stdOut, WErr: stdErr, AppName: AppName, AppVersion: AppVersion, GitCommit: GitCommit},
		},
	}
	sort.Sort(c)
	return c
}

func (c Commander) Len() int {
	return len(c.Commands)
}

func (c Commander) Less(i, j int) bool {
	return c.Commands[i].Key() < c.Commands[j].Key()
}

func (c Commander) Swap(i, j int) {
	c.Commands[i], c.Commands[j] = c.Commands[j], c.Commands[i]
}

// Add will insert a new Command and then sort it by the commands Name field.
func (c *Commander) Add(cmd Command) {
	c.Commands = append(c.Commands, cmd)
	sort.Sort(c)
}

// List will space seperate the list of possible commands.
func (c *Commander) List(writer io.Writer) {
	names := []string{}
	for _, cmd := range c.Commands {
		for k := range cmd.Aliases() {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	fmt.Fprintf(writer, strings.Join(names, " "))
}

// Usage displays the available commands.
func (c *Commander) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "Usage: %s <command>\n\nAvailable commands are:\n", AppName)
	maxLength := 0
	metas := []*nameDescription{}
	for _, cmd := range c.Commands {
		names := []string{}
		for k := range cmd.Aliases() {
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
