package make

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/upsight/ron/color"
)

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

func init() {
	testTargetEnv, testTargetEnvOutput = createTestEnv(nil)
}

func createTestTargetConfig(stdOut *bytes.Buffer, stdErr *bytes.Buffer) (*TargetConfig, *bytes.Buffer, *bytes.Buffer) {
	if stdOut == nil {
		stdOut = &bytes.Buffer{}
	}
	if stdErr == nil {
		stdErr = &bytes.Buffer{}
	}

	tgConf, _ := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargetConfig},
		&Config{Targets: testNewTargets},
	}, stdOut, stdErr)
	return tgConf, stdOut, stdErr
}

func createTestTarget(name string, stdOut *bytes.Buffer, stdErr *bytes.Buffer) (*Target, *bytes.Buffer, *bytes.Buffer) {
	if stdOut == nil {
		stdOut = &bytes.Buffer{}
	}
	if stdErr == nil {
		stdErr = &bytes.Buffer{}
	}

	tgConf, _ := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargetConfig},
		&Config{Targets: testNewTargets},
	}, stdOut, stdErr)
	target, _ := tgConf.Target(name)
	return target, stdOut, stdErr
}

func TestMakeNewTargetConfig(t *testing.T) {
	_, err := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargetConfig},
		&Config{Targets: testNewTargets},
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMakeNewTargetConfigBadDefault(t *testing.T) {
	stdOut := &bytes.Buffer{}
	_, err := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: `:"`},
		&Config{Targets: testNewTargets},
	}, stdOut, nil)
	if err == nil {
		t.Fatal("expected err for invalid default config")
	}
}

func TestMakeNewTargetConfigBadNew(t *testing.T) {
	_, err := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargetConfig},
		&Config{Targets: `:"`},
	}, nil, nil)
	if err == nil {
		t.Fatal("expected err for invalid new config")
	}
}

func TestMakeNewTargetConfigList(t *testing.T) {
	stdOut := &bytes.Buffer{}
	tc, err := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargetConfig},
		&Config{Targets: testNewTargets},
	}, stdOut, nil)
	if err != nil {
		t.Fatal(err)
	}
	tc.List(false, "")
	if !strings.Contains(stdOut.String(), "hello description\n") {
		t.Fatalf("expected hello with description in list of targets, got %s", stdOut.String())
	}
	if !strings.Contains(stdOut.String(), "prep description") {
		t.Fatalf("expected prep in list of targets, got %s", stdOut.String())
	}
}

func TestMakeNewTargetConfigListVerbose(t *testing.T) {
	stdOut := &bytes.Buffer{}
	tc, err := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargetConfig},
		&Config{Targets: testNewTargets},
	}, stdOut, nil)
	if err != nil {
		t.Fatal(err)
	}
	tc.List(true, "")
	if !strings.Contains(stdOut.String(), "hello description") {
		t.Fatalf("expected list of targets, got %s", stdOut.String())
	}
}

func TestMakeNewTargetConfigListVerboseFuzzyGlobbing(t *testing.T) {
	stdOut := &bytes.Buffer{}
	tc, err := NewTargetConfig(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargetConfig},
		&Config{Targets: testNewTargets},
	}, stdOut, nil)
	if err != nil {
		t.Fatal(err)
	}
	tc.List(true, "docker*")
	if !strings.Contains(stdOut.String(), "docker_stats") {
		t.Errorf("expected to get docker_stats in targets")
	}
	stdOut.Reset()
	tc.List(true, "docker_c*")
	if !strings.Contains(stdOut.String(), "docker_clean") {
		t.Errorf("expected to get docker_stats in targets")
	}
	if strings.Contains(stdOut.String(), "docker_stats") {
		t.Errorf("expected to filter out other docker targets")
	}
}

func TestMakeTargetRun(t *testing.T) {
	target, writer, _ := createTestTarget("prep", nil, nil)
	_, _, err := target.Run()
	if err != nil {
		t.Fatal(err)
	}
	want := "hello\nprep1\nprep2\nprep3\nprep4\ngoodbye"
	if !strings.Contains(writer.String(), want) {
		t.Errorf("unexpected output want %s got %s", want, writer.String())
	}
}

func TestMakeTargetRunShellExec(t *testing.T) {
	target, writer, _ := createTestTarget("shellExec", nil, nil)
	_, _, err := target.Run()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(writer.String(), "test") {
		t.Errorf(`cmd not executed want "test" got %s`, writer.String())
	}
}

func TestMakeTargetRunBeforeErr(t *testing.T) {
	target, _, _ := createTestTarget("prepBeforeErr", nil, nil)
	status, _, _ := target.Run()
	if status == 0 {
		t.Fatal("expected non 0 exit status on prep before")
	}
}

func TestMakeTargetRunAfterErr(t *testing.T) {
	target, _, _ := createTestTarget("prepAfterErr", nil, nil)
	status, _, _ := target.Run()
	if status == 0 {
		t.Fatal("expected non 0 exit status on prep after")
	}
}

func TestMakeTargetList(t *testing.T) {
	target, stdOut, _ := createTestTarget("prep", nil, nil)
	target.Description = "description"
	target.List(false, 0)
	want := color.Yellow("prep") + " description\n"
	if stdOut.String() != want {
		t.Errorf("want %s got: %s", want, stdOut.String())
	}
	stdOut.Reset()
	target.List(false, 10)
	want = color.Yellow("prep") + "       description\n"
	if stdOut.String() != want {
		t.Errorf("want %s got: %s", want, stdOut.String())
	}
}

func TestMakeTargetListVerbose(t *testing.T) {
	target, stdOut, _ := createTestTarget("prep", nil, nil)
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
	target, _, _ := createTestTarget("prep", nil, nil)
	target.W = &badWriter{}
	target.List(true, 0)
}
