package target

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/upsight/ron/target"
)

var (
	mockLoadConfig = func(path string) (*target.RawConfig, error) {
		return nil, nil
	}
	mockLoadConfigErr = func(path string) (*target.RawConfig, error) {
		return nil, fmt.Errorf("bad config")
	}
)

func TestRonRunTarget(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{"--yaml=" + path.Join("testdata", "target_test.yaml"), "prep"})
	if err != nil {
		t.Fatal(err)
	}
	if status != 0 {
		t.Fatalf(`expected 0 got %d`, status)
	}
	if stdOut.String() != "hello\nprep\ngoodbye\n" {
		t.Errorf("expected hello\nprep\ngoodbye got %s", stdOut.String())
	}
}

func TestRonRunTargetNArgs(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{})

	if status == 0 {
		t.Fatalf("expected status non 0 got status: %d err: %v", status, err)
	}
}

func TestRonRunTargetListEnvs(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{"--envs"})

	if status != 0 {
		t.Fatalf("expected status 0 got %d %+v", status, err)
	}
	if !strings.Contains(stdOut.String(), "UNAME") {
		t.Errorf("expected UNAME in output got %s", stdOut.String())
	}
}

func TestRonRunConfigsList(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{"--list"})

	if err != nil {
		t.Fatal(err)
	}
	if status != 0 {
		t.Fatalf("expected status 0 got %d", status)
	}
	if !strings.Contains(stdOut.String(), "build") {
		t.Errorf("expected build in output got %s", stdOut.String())
	}
}

func TestRonRunTargetLoadDefaultErr(t *testing.T) {
	prevLoadConfig := target.LoadConfigFile
	defer func() { target.LoadConfigFile = prevLoadConfig }()
	target.LoadConfigFile = mockLoadConfigErr

	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{"--default=" + path.Join("testdata", "target_test.yaml"), "prep"})

	if status != 1 {
		t.Fatalf("expected status 1 got %d", status)
	}
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestRonRunTargetLoadOverrideErr(t *testing.T) {
	prevLoadConfig := target.LoadConfigFile
	defer func() { target.LoadConfigFile = prevLoadConfig }()
	target.LoadConfigFile = mockLoadConfigErr

	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{"--yaml=" + path.Join("testdata", "target_test.yaml"), "prep"})

	if status != 1 {
		t.Fatalf("expected status 1 got %d", status)
	}
	if err == nil {
		t.Fatal("expected err")
	}
}
