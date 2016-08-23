package ron

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	vrs "github.com/upsight/ron/commands/version"
)

func versionCmd(c Commander) *vrs.Command {
	for _, cmd := range c {
		for name := range cmd.Names() {
			if name == "version" {
				v, _ := cmd.(*vrs.Command)
				return v
			}
		}
	}
	return nil
}

func TestNewCommanderUsesStdout(t *testing.T) {
	c := NewCommander(nil, nil)
	v := versionCmd(c)
	if v.W != os.Stdout {
		t.Errorf("writer not set want %+v got %s", os.Stdout, v.W)
	}
}

func TestRonCommanderRunVersionCommand(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := NewCommander(stdOut, stdErr)
	v := versionCmd(c)
	v.Run([]string{})
	got := stdOut.String()
	want := fmt.Sprintln(AppName, AppVersion, GitCommit)
	if got != want {
		t.Errorf("commands run version got %s want %s", got, want)
	}
}
