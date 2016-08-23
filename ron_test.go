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
	status, err := Run(stdOut, stdErr, []string{"version"})
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
	status, _ := Run(stdOut, stdErr, []string{})
	if status != 1 {
		t.Fatalf("expected status 1 got %d", status)
	}
}

func TestRonRunOutputStdout(t *testing.T) {
	status, err := Run(nil, nil, []string{"version"})
	if status != 0 {
		t.Fatalf("expected status 0 got %d", status)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestRonRunNoCommand(t *testing.T) {
	status, _ := Run(nil, nil, []string{"nothere"})
	if status != 1 {
		t.Fatalf("expected status 1 got %d", status)
	}
}
