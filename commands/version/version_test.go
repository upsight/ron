package version

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRonRunVersion(t *testing.T) {
	args := []string{}
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil, AppName: "a", AppVersion: "b", GitCommit: "c"}
	c.Run(args)
	want := fmt.Sprintf("%s %s %s\n", c.AppName, c.AppVersion, c.GitCommit)
	if stdOut.String() != want {
		t.Fatalf("version command want %s got %s", want, stdOut.String())
	}
}
