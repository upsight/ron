package template

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

type badWriter struct {
}

func (bw badWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("bad")
}

// TestNewTemplate verifies attributes are set correctly
func TestTemplateNewTemplate(t *testing.T) {
	var contents = "{{ var }}"
	tmpl := NewTemplate("yaml", contents, nil)
	if tmpl.contents != contents {
		t.Errorf("contents attribute value '%s' does not equal '%s'.", tmpl.contents, contents)
	}
	// When nil it should initialize W to os.Stdout
	if tmpl.Writer != os.Stdout {
		t.Errorf("W attribute should have been 'os.Stdout', got '%s'", tmpl.Writer)
	}
}

// TestRenderBadGoTemplate verifies that if we have bad template syntax
// that we get an error.
func TestTemplateRenderGoBadTemplate(t *testing.T) {
	var writer = &bytes.Buffer{}
	var contents = "Hi, {{ foo }, this is a bad templated"
	tmpl := NewTemplate("yaml", contents, writer)
	if err := tmpl.Render(); err == nil {
		t.Error("Bad template did not return an error.")
	}
}

// TestRender tests that template variables are replaced with the matching
// environment variables, or the default if the default filter is used.
func TestTemplateRenderGo(t *testing.T) {
	var writer = &bytes.Buffer{}
	var contents = `{{define "bob"}}bobobla{{end}}Hi, {{.Env.FOO}}, {{.Env.doo}} my name is {{template "bob"}}`
	os.Setenv("FOO", "foo===foo")
	os.Setenv("doo", "doodoo")
	tmpl := NewTemplate("yaml", contents, writer)
	tmpl.Render()
	var want = "Hi, foo===foo, doodoo my name is bobobla"
	if writer.String() != want {
		t.Errorf("Rendered template string didn't match '%s', got '%s'", want, writer.String())
	}
}

// TestRenderBadMacro verifies that if we use an incomplete exported macro
// that we will get an error.  macro here requires an argument which we
// are not passing.
func TestTemplateRenderBadMacro(t *testing.T) {
	var writer = &bytes.Buffer{}
	var contents = `{{Hi, {% macro %}, this is a bad templated`
	tmpl := NewTemplate("yaml", contents, writer)
	if err := tmpl.Render(); err == nil {
		t.Error("Bad context var did not return an error.")
	}
}

// TestRenderCantWriteString varifies we get an error if we can't write
// to the io.Writer.
func TestTemplateRenderCantWriteString(t *testing.T) {
	var writer = &badWriter{}
	var contents = "{{ var }}"
	tmpl := NewTemplate("yaml", contents, writer)
	if err := tmpl.Render(); err == nil {
		t.Error("Expected badWriter to fail to write, but it succeeded")
	}
}
