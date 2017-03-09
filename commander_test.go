package ron

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	vrs "github.com/upsight/ron/commands/version"
)

func versionCmd(c *Commander) *vrs.Command {
	for _, cmd := range c.Commands {
		for name := range cmd.Aliases() {
			if name == "version" {
				v, _ := cmd.(*vrs.Command)
				return v
			}
		}
	}
	return nil
}

func TestNewDefaultCommanderUsesStdout(t *testing.T) {
	c := NewDefaultCommander(nil, nil)
	v := versionCmd(c)
	if v.W != os.Stdout {
		t.Errorf("writer not set want %+v got %s", os.Stdout, v.W)
	}
}

func TestRonCommanderRunVersionCommand(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := NewDefaultCommander(stdOut, stdErr)
	v := versionCmd(c)
	v.Run([]string{})
	got := stdOut.String()
	want := fmt.Sprintln(AppName, AppVersion, GitCommit)
	if got != want {
		t.Errorf("commands run version got %s want %s", got, want)
	}
}

func TestRonCommanderList(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := NewDefaultCommander(stdOut, stdErr)
	c.List(stdOut)
	want := "bash_completion burgundy cmd hs httpstat replace t target template upgrade version"
	if stdOut.String() != want {
		t.Errorf("want %q, got %q", want, stdOut.String())
	}
}

type TestCommand struct {
	Name string
}

func (c *TestCommand) Run(args []string) (int, error) {
	return 0, nil
}

func (c *TestCommand) Key() string {
	return c.Name
}

func (c *TestCommand) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"test": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *TestCommand) Description() string {
	return "Print the version."
}

func TestRonCommanderAdd(t *testing.T) {
	c := NewDefaultCommander(nil, nil)
	c.Add(&TestCommand{Name: "apples"})
	c.Add(&TestCommand{Name: "bpples"})
	if c.Commands[0].Key() != "apples" {
		t.Errorf("want apples, got %s", c.Commands[0].Key())
	}
	if c.Commands[2].Key() != "bpples" {
		t.Errorf("want bpples, got %s", c.Commands[0].Key())
	}
}
