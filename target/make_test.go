package target

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var (
	wrkdir string
)

func init() {
	log.SetOutput(ioutil.Discard)
	wrkdir, _ = os.Getwd()
}

type badWriter struct {
}

func (bw badWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("bad")
}

func TestNewMake(t *testing.T) {
	tc, _, _ := createTestConfigs(t, nil, nil)
	_, err := NewMake(tc)
	ok(t, err)
}

func TestMakeRun(t *testing.T) {
	tc, tcW, _ := createTestConfigs(t, nil, nil)
	m, _ := NewMake(tc)
	err := m.Run("ron:prep")
	ok(t, err)
	want := "hello\nprep1\nprep2\nprep3\nprep4\ngoodbye\n"
	equals(t, want, tcW.String())
}

func TestMakeRunErr(t *testing.T) {
	tc, _, _ := createTestConfigs(t, nil, nil)
	m, _ := NewMake(tc)
	err := m.Run("err")
	if err == nil {
		t.Fatal("expected target not found")
	}
}

func TestMakeRunNoTarget(t *testing.T) {
	tc, _, _ := createTestConfigs(t, nil, nil)
	m, _ := NewMake(tc)
	err := m.Run("prop")
	if err == nil {
		t.Fatal("expected target not found")
	}
}
