package target

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"testing"

	mke "github.com/upsight/ron/make"
)

var (
	mockLoadConfig = func(path string) (string, string, error) {
		return "", "", nil
	}
	mockLoadConfigErr = func(path string) (string, string, error) {
		return "", "", fmt.Errorf("bad config")
	}
)

func TestRonRunTarget(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{"--default=" + path.Join("testdata", "target_test.yaml"), "prep"})
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

func TestRonRunTargetConfigList(t *testing.T) {
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
	prevLoadConfig := mke.LoadConfigFile
	defer func() { mke.LoadConfigFile = prevLoadConfig }()
	mke.LoadConfigFile = mockLoadConfigErr

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
	prevLoadConfig := mke.LoadConfigFile
	defer func() { mke.LoadConfigFile = prevLoadConfig }()
	mke.LoadConfigFile = mockLoadConfigErr

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
