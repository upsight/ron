package make

import (
	"bytes"
	"log"
	"reflect"
	"strings"
	"testing"
)

type parseOSEnvsTest struct {
	in  []string
	out MSS
}

var testNewEnvConfig = `
- UNAME: plan9
- RON: was here
- CMD: +echo blah
- ENVS: >-
    -e CMD=$CMD
    -e TEST=$UNAME
- NOOP:
`

func createTestEnv(writer *bytes.Buffer) (*Env, *bytes.Buffer) {
	if writer == nil {
		writer = &bytes.Buffer{}
	}
	e, err := NewEnv(DefaultEnvConfig, "", MSS{"UNAME": "plan9"}, writer)
	if err != nil {
		log.Fatal(err)
	}
	return e, writer
}

func TestMakeNewEnv(t *testing.T) {
	writer := &bytes.Buffer{}
	NewEnv(DefaultEnvConfig, "", MSS{}, writer)
}

func TestMakeNewEnvStdout(t *testing.T) {
	NewEnv(DefaultEnvConfig, "", MSS{}, nil)
}

func TestMakeIn(t *testing.T) {
	type inTest struct {
		inKey  string
		inKeys []string
		out    bool
	}

	var inTests = []inTest{
		{"a", []string{"a", "b", "c"}, true},
		{"x", []string{"a", "b", "c"}, false},
		{"a", []string{"b", "z", "y", "a"}, true},
		{"a", []string{"b", "z", "a", "y"}, true},
		{"x", []string{}, false},
		{"", []string{}, false},
	}
	for i, test := range inTests {
		got := in(test.inKey, test.inKeys)
		if got != test.out {
			t.Fatalf("%d want %+v got %+v", i, test.out, got)
		}
	}
}

func TestMakeParseOSEnvs(t *testing.T) {
	var parseOSEnvTests = []parseOSEnvsTest{
		{[]string{"a="}, MSS{"a": ""}},
		{[]string{"a=b"}, MSS{"a": "b"}},
		{[]string{"a=b = 1"}, MSS{"a": "b = 1"}},
		{[]string{"b=> &>1"}, MSS{"b": "> &>1"}},
	}
	for i, test := range parseOSEnvTests {
		got := ParseOSEnvs(test.in)
		if !reflect.DeepEqual(got, test.out) {
			t.Fatalf("%d want %+v got %+v", i, test.out, got)
		}
	}
}

func TestMakeEnvProcess(t *testing.T) {
	writer := &bytes.Buffer{}
	e, err := NewEnv(DefaultEnvConfig, testNewEnvConfig, ParseOSEnvs([]string{"HELLO=hello", "ABC=+pwd"}), writer)
	if err != nil {
		t.Fatal(err)
	}

	got := e.Config["HELLO"]
	want := `hello`
	if !strings.Contains(got, want) {
		t.Fatalf("config does not contain %s got \n%s", want, got)
	}

	got = e.Config["UNAME"]
	want = `plan9`
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}

	got = e.Config["APP_UNDERSCORE"]
	want = "ron"
	if got != want {
		for _, k := range e.keyOrder {
			t.Log(k, e.Config[k])
		}
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestMakeEnvProcessEnv(t *testing.T) {
	writer := &bytes.Buffer{}
	e, err := NewEnv(DefaultEnvConfig, testNewEnvConfig, ParseOSEnvs([]string{}), writer)
	if err != nil {
		t.Fatal(err)
	}

	got := e.Config["ENVS"]
	want := `-e CMD=blah -e TEST=plan9`
	if !strings.Contains(got, want) {
		t.Fatalf("config ENVS does not contain %s got \n%s", want, got)
	}
}

func TestMakeEnvProcessBadCommand(t *testing.T) {
	writer := &bytes.Buffer{}
	_, err := NewEnv(DefaultEnvConfig, testNewEnvConfig+"\nHELLO=+hello", ParseOSEnvs([]string{}), writer)
	if err == nil {
		t.Fatal("expected err processing command +hello")
	}
}

func TestMakeEnvProcessBadYaml(t *testing.T) {
	_, err := NewEnv(`:"`, "", MSS{}, nil)
	if err == nil {
		t.Fatal("should have gotten invalid err")
	}
}

func TestMakeEnvProcessBadYamlNewEnvs(t *testing.T) {
	_, err := NewEnv("", `:"`, MSS{}, nil)
	if err == nil {
		t.Fatal("should have gotten invalid err")
	}
}

func TestMakeEnvList(t *testing.T) {
	writer := &bytes.Buffer{}
	e, _ := NewEnv(DefaultEnvConfig, "", MSS{}, writer)
	e.List()
	got := writer.String()
	want := "ron\n"
	if !strings.Contains(got, want) {
		t.Fatalf("output does not contain %s got \n%s", want, got)
	}
}

func TestMakeEnvListBadWriter(t *testing.T) {
	e, _ := NewEnv(DefaultEnvConfig, "", MSS{}, badWriter{})
	e.List()
}

func TestMakeEnvPrintRaw(t *testing.T) {
	writer := &bytes.Buffer{}
	e, _ := NewEnv(DefaultEnvConfig, "", MSS{}, writer)
	err := e.PrintRaw()
	if err != nil {
		t.Fatal(err)
	}
	got := writer.String()
	if got != DefaultEnvConfig+"\n\n" {
		t.Fatalf("output does not match raw yaml %s", got)
	}
}

func TestMakeEnvPrintRawBadWriter(t *testing.T) {
	e, err := NewEnv(DefaultEnvConfig, "", MSS{}, badWriter{})
	if err != nil {
		t.Fatal(err)
	}
	e.PrintRaw()
}
