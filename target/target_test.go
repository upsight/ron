package target

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/upsight/ron/color"
)

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestTargetRun(t *testing.T) {
	target, writer, _ := createTestTarget(t, "ron:prep", nil, nil)
	_, _, err := target.Run()
	ok(t, err)
	want := "hello\nprep1\nprep2\nprep3\nprep4\ngoodbye"
	if !strings.Contains(writer.String(), want) {
		t.Errorf("unexpected output want %q got %q", want, writer.String())
	}
}

func TestTargetRunShellExec(t *testing.T) {
	target, writer, _ := createTestTarget(t, "shellExec", nil, nil)
	_, _, err := target.Run()
	ok(t, err)
	if !strings.Contains(writer.String(), "test") {
		t.Errorf(`cmd not executed want "test" got %q`, writer.String())
	}
}

func TestTargetRunBeforeErr(t *testing.T) {
	target, _, _ := createTestTarget(t, "prepBeforeErr", nil, nil)
	status, _, _ := target.Run()
	if status == 0 {
		t.Fatal("expected non 0 exit status on prep before")
	}
}

func TestTargetRunAfterErr(t *testing.T) {
	target, _, _ := createTestTarget(t, "prepAfterErr", nil, nil)
	status, _, _ := target.Run()
	if status == 0 {
		t.Fatal("expected non 0 exit status on prep after")
	}
}

func TestTargetList(t *testing.T) {
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

func TestTargetListVerbose(t *testing.T) {
	target, stdOut, _ := createTestTarget(t, "ron:prep", nil, nil)
	target.List(true, 0)
	want := "before: hello, prep"
	if !strings.Contains(stdOut.String(), want) {
		t.Errorf(`want in string %q got %q`, want, stdOut.String())
	}
	want = "after: goodbye, prep"
	if !strings.Contains(stdOut.String(), want) {
		t.Errorf(`want in string %q got %q`, want, stdOut.String())
	}
	want = "prep4"
	if !strings.Contains(stdOut.String(), want) {
		t.Errorf(`want in string %q got %q`, want, stdOut.String())
	}
}

func TestTargetListBadWriter(t *testing.T) {
	target, _, _ := createTestTarget(t, "prep", nil, nil)
	target.W = &badWriter{}
	target.List(true, 0)
}
