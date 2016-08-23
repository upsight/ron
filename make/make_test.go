package make

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
	e, _ := createTestEnv(nil)
	tc, _, _ := createTestTargetConfig(nil, nil)
	_, err := NewMake(e, tc)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMakeRun(t *testing.T) {
	e, _ := createTestEnv(nil)
	tc, tcW, _ := createTestTargetConfig(nil, nil)
	m, _ := NewMake(e, tc)
	err := m.Run("prep")
	if err != nil {
		t.Fatal(err)
	}
	want := "hello\nprep1\nprep2\nprep3\nprep4\ngoodbye\n"
	if tcW.String() != want {
		t.Fatalf("expected %s got %s", want, tcW.String())
	}

	// test env substitution
	tcW.Reset()
	err = m.Run("uname")
	if err != nil {
		t.Fatal(err)
	}
	if tcW.String() != "plan9\n" {
		t.Fatalf("expected plan9 got %s", tcW.String())
	}
}

func TestMakeRunErr(t *testing.T) {
	e, _ := createTestEnv(nil)
	tc, _, _ := createTestTargetConfig(nil, nil)
	m, _ := NewMake(e, tc)
	err := m.Run("err")
	if err == nil {
		t.Fatal("expected target not found")
	}
}

func TestMakeRunNoTarget(t *testing.T) {
	e, _ := createTestEnv(nil)
	tc, _, _ := createTestTargetConfig(nil, nil)
	m, _ := NewMake(e, tc)
	err := m.Run("prop")
	if err == nil {
		t.Fatal("expected target not found")
	}
}
