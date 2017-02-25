package make

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/upsight/ron/color"
)

//testTargetEnv, testTargetEnvOutput = createTestEnv(nil)

var (
	testTargetEnv       *Env
	testTargetEnvOutput io.Writer
	testNewTargets      = `
hello:
  description: hello description
  cmd: |
    echo "hello"
goodbye:
  description: goodbye description
  cmd: |
    echo "goodbye"
prep:
  description: prep description
  before:
    - hello
    - prep
  after:
    - goodbye
    - prep
  cmd: |
    echo "prep1"
    echo "prep2" && \
      echo "prep3"
    if [ 1 != 0 ]; then \
      echo "prep4"
    else \
      echo "prepnope"; \
    fi
uname:
  cmd: |
    echo $UNAME
shellExec:
  cmd: |
    echo $(echo test)
err:
  cmd: |
    me_garbage
prepBeforeErr:
  before:
    - err
  after:
    - goodbye
  cmd: |
    echo 1
prepAfterErr:
  before:
    - hello
  after:
    - err
  cmd: |
    echo 1
`
)

func createTestTarget(t *testing.T, name string, stdOut *bytes.Buffer, stdErr *bytes.Buffer) (*Target, *bytes.Buffer, *bytes.Buffer) {
	if stdOut == nil {
		stdOut = &bytes.Buffer{}
	}
	if stdErr == nil {
		stdErr = &bytes.Buffer{}
	}

	testTargetEnv, _ = createTestEnv(t, nil)
	tgConf, _ := NewTargetConfigs(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargets},
		&Config{Targets: testNewTargets},
	}, stdOut, stdErr)
	target, _ := tgConf.Target(name)
	return target, stdOut, stdErr
}

func TestMakeTargetRun(t *testing.T) {
	target, writer, _ := createTestTarget(t, "prep", nil, nil)
	_, _, err := target.Run()
	ok(t, err)
	want := "hello\nprep1\nprep2\nprep3\nprep4\ngoodbye"
	if !strings.Contains(writer.String(), want) {
		t.Errorf("unexpected output want %s got %s", want, writer.String())
	}
}

func TestMakeTargetRunShellExec(t *testing.T) {
	target, writer, _ := createTestTarget(t, "shellExec", nil, nil)
	_, _, err := target.Run()
	ok(t, err)
	if !strings.Contains(writer.String(), "test") {
		t.Errorf(`cmd not executed want "test" got %s`, writer.String())
	}
}

func TestMakeTargetRunBeforeErr(t *testing.T) {
	target, _, _ := createTestTarget(t, "prepBeforeErr", nil, nil)
	status, _, _ := target.Run()
	if status == 0 {
		t.Fatal("expected non 0 exit status on prep before")
	}
}

func TestMakeTargetRunAfterErr(t *testing.T) {
	target, _, _ := createTestTarget(t, "prepAfterErr", nil, nil)
	status, _, _ := target.Run()
	if status == 0 {
		t.Fatal("expected non 0 exit status on prep after")
	}
}

func TestMakeTargetList(t *testing.T) {
	target, stdOut, _ := createTestTarget(t, "prep", nil, nil)
	target.Description = "description"
	target.List(false, 0)
	want := color.Yellow("prep") + " description\n"
	equals(t, want, stdOut.String())
	stdOut.Reset()
	target.List(false, 10)
	want = color.Yellow("prep") + "       description\n"
	equals(t, want, stdOut.String())
}

func TestMakeTargetListVerbose(t *testing.T) {
	target, stdOut, _ := createTestTarget(t, "prep", nil, nil)
	target.List(true, 0)
	want := "before: hello, prep"
	if !strings.Contains(stdOut.String(), want) {
		t.Errorf(`want in string "%s" got "%s"`, want, stdOut.String())
	}
	want = "after: goodbye, prep"
	if !strings.Contains(stdOut.String(), want) {
		t.Errorf(`want in string "%s" got "%s"`, want, stdOut.String())
	}
	want = "prep4"
	if !strings.Contains(stdOut.String(), want) {
		t.Errorf(`want in string "%s" got "%s"`, want, stdOut.String())
	}
}

func TestMakeTargetListBadWriter(t *testing.T) {
	target, _, _ := createTestTarget(t, "prep", nil, nil)
	target.W = &badWriter{}
	target.List(true, 0)
}
