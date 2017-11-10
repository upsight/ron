package template

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func makeTempFile(t *testing.T, s string) string {
	f, err := ioutil.TempFile(os.TempDir(), "template_test")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if s != "" {
		_, err := f.WriteString(s)
		if err != nil {
			os.Remove(f.Name())
			t.Fatal(err)
		}
	}
	return f.Name()
}

func TestRonRunTemplate(t *testing.T) {
	os.Setenv("FOO", "xyz")
	inPath := makeTempFile(t, "a {{ .Env.FOO }} b")
	defer os.Remove(inPath)
	outPath := makeTempFile(t, "")
	defer os.Remove(outPath)

	stdErr := &bytes.Buffer{}
	c := &Command{W: nil, WErr: stdErr}
	status, err := c.Run([]string{"-input", inPath, "-output", outPath})
	if err != nil {
		t.Fatal(err)
	}
	if status != 0 {
		t.Fatalf("expected status 0 got %d", status)
	}

	b, err := ioutil.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	expected := "a xyz b"
	if string(b) != expected {
		t.Fatalf(`expected "%s", got "%s"`, expected, string(b))
	}
	if stdErr.String() != "" {
		t.Fatalf(`expected no output, got "%s"`, stdErr.String())
	}
}

func TestRonRunTemplateStdout(t *testing.T) {
	os.Setenv("FOO", "xyz")
	inPath := makeTempFile(t, "a {{ .Env.FOO }} b")

	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, err := c.Run([]string{"-input", inPath})

	if err != nil {
		t.Fatal(err)
	}
	if status != 0 {
		t.Fatalf("expected status 0, got %d", status)
	}

	expected := "a xyz b"
	if stdOut.String() != expected {
		t.Fatalf(`expected "%s", got "%s"`, expected, stdOut.String())
	}
}

func TestRonRunTemplateErrorInput(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"-input", ""})

	if status != 1 {
		t.Fatalf("expected status 1, got %d", status)
	}
}

func TestRonRunTemplateErrorNoFile(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"-input", "/bad/template/filename"})

	if status != 1 {
		t.Fatalf("expected status 1, got %d", status)
	}
}

func TestRonRunTemplateErrorCreateOutFile(t *testing.T) {
	inPath := makeTempFile(t, "a {{ bad template } b")
	defer os.Remove(inPath)

	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"--input", inPath, "--output", "/dev/does/not/exist"})

	if status != 1 {
		t.Fatalf("expected status 1, got %d", status)
	}
}

func TestRonRunTemplateErrorExport(t *testing.T) {
	inPath := makeTempFile(t, "a {{ bad template } b")
	defer os.Remove(inPath)

	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"--input", inPath})

	if status != 1 {
		t.Fatalf("expected status 1, got %d", status)
	}
}
