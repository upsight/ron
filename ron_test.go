package ron

import (
	"bytes"
	"os"
	"testing"
)

var (
	wrkdir string
)

func init() {
	wrkdir, _ = os.Getwd()
}

func TestRonRun(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := NewDefaultCommander(stdOut, stdErr)
	status, err := Run(c, []string{"version"})
	if status != 0 {
		t.Fatalf("expected status 0 got %d", status)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestRonRunNArgs(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := NewDefaultCommander(stdOut, stdErr)
	status, _ := Run(c, []string{})
	if status != 1 {
		t.Fatalf("expected status 1 got %d", status)
	}
}

func TestRonRunOutputStdout(t *testing.T) {
	c := NewDefaultCommander(nil, nil)
	status, err := Run(c, []string{"version"})
	if status != 0 {
		t.Fatalf("expected status 0 got %d", status)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestRonRunNoCommand(t *testing.T) {
	c := NewDefaultCommander(nil, nil)
	status, _ := Run(c, []string{"nothere"})
	if status != 1 {
		t.Fatalf("expected status 1 got %d", status)
	}
}
