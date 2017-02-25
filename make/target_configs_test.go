package make

import (
	"bytes"
	"strings"
	"testing"
)

func createTestTargetConfigs(t *testing.T, stdOut *bytes.Buffer, stdErr *bytes.Buffer) (*TargetConfigs, *bytes.Buffer, *bytes.Buffer) {
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
	return tgConf, stdOut, stdErr
}

func TestMakeNewTargetConfigsListVerboseFuzzyGlobbing(t *testing.T) {
	stdOut := &bytes.Buffer{}
	testTargetEnv, _ := createTestEnv(t, nil)
	tc, err := NewTargetConfigs(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargets, Filepath: "default.yaml"},
		&Config{Targets: testNewTargets, Filepath: "test.yaml"},
	}, stdOut, nil)
	if err != nil {
		t.Fatal(err)
	}
	tc.List(true, "docker*")
	if !strings.Contains(stdOut.String(), "docker_stats") {
		t.Errorf("expected to get docker_stats in targets, got %q", stdOut.String())
	}
	stdOut.Reset()
	tc.List(true, "docker_c*")
	if !strings.Contains(stdOut.String(), "docker_clean") {
		t.Errorf("expected to get docker_stats in targets, got %q", stdOut.String())
	}
	if strings.Contains(stdOut.String(), "docker_stats") {
		t.Errorf("expected to filter out other docker targets, got %q", stdOut.String())
	}
}

func TestMakeNewTargetConfigsListVerbose(t *testing.T) {
	stdOut := &bytes.Buffer{}
	testTargetEnv, _ := createTestEnv(t, nil)
	tc, err := NewTargetConfigs(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargets},
		&Config{Targets: testNewTargets},
	}, stdOut, nil)
	ok(t, err)
	tc.List(true, "")
	if !strings.Contains(stdOut.String(), "hello description") {
		t.Fatalf("expected list of targets, got %q", stdOut.String())
	}
}

func TestMakeNewTargetConfigsList(t *testing.T) {
	stdOut := &bytes.Buffer{}
	testTargetEnv, _ := createTestEnv(t, nil)
	tc, err := NewTargetConfigs(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargets},
		&Config{Targets: testNewTargets},
	}, stdOut, nil)
	if err != nil {
		t.Fatal(err)
	}
	tc.List(false, "")
	if !strings.Contains(stdOut.String(), "hello description\n") {
		t.Fatalf("expected hello with description in list of targets, got %q", stdOut.String())
	}
	if !strings.Contains(stdOut.String(), "prep description") {
		t.Fatalf("expected prep in list of targets, got %q", stdOut.String())
	}
}

func TestMakeNewTargetConfigsBadDefault(t *testing.T) {
	stdOut := &bytes.Buffer{}
	testTargetEnv, _ := createTestEnv(t, nil)
	_, err := NewTargetConfigs(testTargetEnv, []*Config{
		&Config{Targets: `:"`},
		&Config{Targets: testNewTargets},
	}, stdOut, nil)
	if err == nil {
		t.Fatal("expected err for invalid default config")
	}
}

func TestMakeNewTargetConfigsBadNew(t *testing.T) {
	testTargetEnv, _ := createTestEnv(t, nil)
	_, err := NewTargetConfigs(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargets},
		&Config{Targets: `:"`},
	}, nil, nil)
	if err == nil {
		t.Fatal("expected err for invalid new config")
	}
}

func TestMakeNewTargetConfigs(t *testing.T) {
	testTargetEnv, _ := createTestEnv(t, nil)
	_, err := NewTargetConfigs(testTargetEnv, []*Config{
		&Config{Targets: DefaultTargets},
		&Config{Targets: testNewTargets},
	}, nil, nil)
	ok(t, err)
}
