package replace

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func checkFile(path, token string, count int) (int, bool) {
	d, _ := ioutil.ReadFile(path)
	c := strings.Count(string(d), token)
	if c != count {
		return c, false
	}
	return c, true
}

func TestRonRunReplaceShouldIgnore(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		{"abc/.git/a", true},
		{"abc/.hg/b", true},
		{"abc/abc.ron.tmp", true},
		{"abc/abc.ron", false},
		{"abc/abc.txt", false},
	}

	for _, tt := range tests {
		got := shouldIgnore(tt.in)
		if got != tt.out {
			t.Errorf("shouldIgnore(%s) => %v, want %v", tt.in, got, tt.out)
		}
	}
}

func TestRonRunReplace(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	if count, ok := checkFile("testdata/test.txt", "ron", 2); !ok {
		t.Errorf("want 2 got %d", count)
	}
	c := &Command{W: stdOut, WErr: stdErr}
	status, err := c.Run([]string{"testdata", "ron", "run"})
	if status != 0 {
		t.Fatalf("expected status 0 got %d", status)
	}
	if err != nil {
		t.Fatal(err)
	}
	if count, ok := checkFile("testdata/test.txt", "ron", 0); !ok {
		t.Errorf("want 0 got %d", count)
	}
	if count, ok := checkFile("testdata/test.txt", "run", 2); !ok {
		t.Errorf("want 2 got %d", count)
	}
	status, err = c.Run([]string{"testdata", "run", "ron"})
	if status != 0 {
		t.Fatalf("expected status 0 got %d", status)
	}
	if err != nil {
		t.Fatal(err)
	}
	if count, ok := checkFile("testdata/test.txt", "ron", 2); !ok {
		t.Errorf("want 2 got %d", count)
	}
}
