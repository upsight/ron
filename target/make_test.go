package target

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

var (
	wrkdir string
)

func init() {
	log.SetOutput(ioutil.Discard)
	wrkdir, _ = os.Getwd()
}

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

type badWriter struct {
}

func (bw badWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("bad")
}

func TestNewMake(t *testing.T) {
	e, _ := createTestEnv(t, nil)
	tc, _, _ := createTestTargetConfigs(t, nil, nil)
	_, err := NewMake(e, tc)
	ok(t, err)
}

func TestMakeRun(t *testing.T) {
	e, _ := createTestEnv(t, nil)
	tc, tcW, _ := createTestTargetConfigs(t, nil, nil)
	m, _ := NewMake(e, tc)
	err := m.Run("prep")
	ok(t, err)
	want := "hello\nprep1\nprep2\nprep3\nprep4\ngoodbye\n"
	equals(t, want, tcW.String())

	// test env substitution
	tcW.Reset()
	err = m.Run("uname")
	ok(t, err)
	equals(t, "plan9\n", tcW.String())
}

func TestMakeRunErr(t *testing.T) {
	e, _ := createTestEnv(t, nil)
	tc, _, _ := createTestTargetConfigs(t, nil, nil)
	m, _ := NewMake(e, tc)
	err := m.Run("err")
	if err == nil {
		t.Fatal("expected target not found")
	}
}

func TestMakeRunNoTarget(t *testing.T) {
	e, _ := createTestEnv(t, nil)
	tc, _, _ := createTestTargetConfigs(t, nil, nil)
	m, _ := NewMake(e, tc)
	err := m.Run("prop")
	if err == nil {
		t.Fatal("expected target not found")
	}
}
