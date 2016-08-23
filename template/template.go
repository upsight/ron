// Package template renders templates to files.
// It attempts to render them as Go text/templates.
package template

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/upsight/ron/color"
)

// Debug toggles debug on or off.
var (
	Debug = false
)

// Template renders Jinja templates to files.
type Template struct {
	name     string
	contents string
	io.Writer
}

// NewTemplate creates a new instance of Template.
func NewTemplate(name, contents string, writer io.Writer) *Template {
	if writer == nil {
		writer = os.Stdout
	}
	return &Template{name, contents, writer}
}

// RenderGo will attempt to render as a Go template.
func RenderGo(name, contents string) (string, error) {
	t := template.New(name).Funcs(template.FuncMap(FuncMap))
	t, err := t.Parse(contents)
	if err != nil {
		return "", err
	}
	b := bytes.Buffer{}
	// Only render as a template if Parse works.
	err = t.ExecuteTemplate(&b, name, NewContext())
	if err != nil {
		return "", err
	}
	contents = b.String()
	return contents, err
}

// Render outputs a templated file using Go text/template.
func (t *Template) Render() error {
	var out []byte
	tmpl, err := RenderGo(t.name, t.contents)
	if err != nil {
		return err
	}
	if Debug {
		fmt.Println(color.Blue(tmpl))
	}
	out = []byte(tmpl)

	_, err = t.Writer.Write(out)
	if err != nil {
		return err
	}
	return nil
}
