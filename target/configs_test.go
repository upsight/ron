package target

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

var (
	defaultTestConfigRaw    string
	ronTestConfigRaw        string
	defaultTargetConfigFile *ConfigFile
	ronTargetConfigFile     *ConfigFile
)

func init() {
	d, _ := ioutil.ReadFile("testdata/default.yaml")
	defaultTestConfigRaw = string(d)
	err := yaml.Unmarshal(d, &defaultTargetConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	d, _ = ioutil.ReadFile("testdata/ron.yaml")
	ronTestConfigRaw = string(d)
	err = yaml.Unmarshal(d, &ronTargetConfigFile)
	if err != nil {
		log.Fatal(err)
	}
}

func createTestEnv(t *testing.T, writer *bytes.Buffer) (*Env, *bytes.Buffer) {
	if writer == nil {
		writer = &bytes.Buffer{}
	}
	envs, _, err := BuiltinDefault()
	ok(t, err)
	e, err := NewEnv(nil, &RawConfig{Envs: envs}, MSS{"UNAME": "plan9"}, writer)
	ok(t, err)
	return e, writer
}

func createTestTarget(t *testing.T, name string, stdOut *bytes.Buffer, stdErr *bytes.Buffer) (*Target, *bytes.Buffer, *bytes.Buffer) {
	tc, stdOut, stdErr := createTestConfigs(t, stdOut, stdErr)
	target, ok := tc.Target(name)
	if !ok {
		t.FailNow()
	}
	return target, stdOut, stdErr
}

func createTestConfigs(t *testing.T, stdOut *bytes.Buffer, stdErr *bytes.Buffer) (*Configs, *bytes.Buffer, *bytes.Buffer) {
	if stdOut == nil {
		stdOut = &bytes.Buffer{}
	}
	if stdErr == nil {
		stdErr = &bytes.Buffer{}
	}

	tc, err := NewConfigs([]*RawConfig{
		&RawConfig{
			Filepath: "testdata/default.yaml",
			Envs:     defaultTargetConfigFile.EnvsString(),
			Targets:  defaultTargetConfigFile.TargetsString(),
		},
		&RawConfig{
			Filepath: "testdata/ron.yaml",
			Envs:     ronTargetConfigFile.EnvsString(),
			Targets:  ronTargetConfigFile.TargetsString(),
		},
	}, "", stdOut, stdErr)
	ok(t, err)
	return tc, stdOut, stdErr
}

func TestNewConfigsListVerboseFuzzyGlobbing(t *testing.T) {
	stdOut := &bytes.Buffer{}
	tc, _, _ := createTestConfigs(t, stdOut, nil)

	tests := []struct {
		name           string
		target         string
		wantStringIn   string // a string that should be returned
		noWantStringIn string // a string that shouldn't match
	}{
		{"00 without filename find the default target", "default", "echo default default", "echo default prep"},
		{"00 without filename find the target in ron", "hello", "echo hello", "abc"},
		{"00 without filename find glob target in ron", "h*", "echo hello", "abc"},
		{"01 with filename find the default target", "default:default", "echo default default", "echo default prep"},
		{"02 with filename find the default target glob", "default:d*", "echo default default", "echo default prep"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.List(true, tt.target)
			if !strings.Contains(stdOut.String(), tt.wantStringIn) {
				t.Errorf("expected to get %q in targets, got %q", tt.wantStringIn, stdOut.String())
			}
			if tt.noWantStringIn != "" {
				if strings.Contains(stdOut.String(), tt.noWantStringIn) {
					t.Errorf("expected to not get %q in targets, got %q", tt.wantStringIn, stdOut.String())
				}
			}
			stdOut.Reset()
		})
	}
}

func TestNewConfigsListVerbose(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	tc, stdOut, stdErr := createTestConfigs(t, stdOut, stdErr)

	tc.List(true, "")
	if !strings.Contains(stdOut.String(), "hello description") {
		t.Fatalf("expected list of targets, got %q", stdOut.String())
	}
}

func TestNewConfigsList(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	tc, stdOut, stdErr := createTestConfigs(t, stdOut, stdErr)

	tc.List(false, "")
	if !strings.Contains(stdOut.String(), "hello description\n") {
		t.Fatalf("expected hello with description in list of targets, got %q", stdOut.String())
	}
	if !strings.Contains(stdOut.String(), "prep description") {
		t.Fatalf("expected prep in list of targets, got %q", stdOut.String())
	}
}

func TestNewConfigsListClean(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	tc, stdOut, stdErr := createTestConfigs(t, stdOut, stdErr)

	tc.ListClean()
	want := "default:default default:echo default:prep ron:echo ron:err ron:goodbye ron:hello ron:prep ron:prepAfterErr ron:prepBeforeErr ron:shellExec ron:uname "
	equals(t, want, stdOut.String())
}

func TestNewConfigsBadDefault(t *testing.T) {
	stdOut := &bytes.Buffer{}
	_, err := NewConfigs([]*RawConfig{&RawConfig{Targets: `:"`}}, "", stdOut, nil)

	if err == nil {
		t.Fatal("expected err for invalid default config")
	}
}

func TestNewConfigsBadNew(t *testing.T) {
	_, err := NewConfigs([]*RawConfig{&RawConfig{Targets: `:"`}}, "", nil, nil)
	if err == nil {
		t.Fatal("expected err for invalid new config")
	}
}
