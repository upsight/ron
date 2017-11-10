package template

import (
	"flag"
	"fmt"
	"io"
	"os"

	fi "github.com/upsight/ron/file"
	"github.com/upsight/ron/template"
)

// Command ...
type Command struct {
	W       io.Writer
	WErr    io.Writer
	AppName string
	Name    string
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	fs := flag.NewFlagSet(c.Name, flag.ExitOnError)
	fs.SetOutput(c.WErr)
	fs.Usage = func() {
		fmt.Fprintf(c.W, "Usage: %s %s\n", c.AppName, c.Name)
		fs.PrintDefaults()
	}

	var inPath string
	var outPath string
	fs.StringVar(&inPath, "input", "", "Path or URL to template file.")
	fs.StringVar(&outPath, "output", "", `Path to output file. Defaults to stdout.`)
	fs.BoolVar(&template.Debug, "debug", false, "Debug the template being run")
	fs.Parse(args)
	if inPath == "" {
		fs.Usage()
		return 1, nil
	}

	f, err := fi.NewFile(inPath)
	if err != nil {
		return 1, err
	}

	var w io.Writer
	switch {
	case outPath == "":
		w = c.W
	default:
		of, err := os.Create(outPath)
		if err != nil {
			return 1, err
		}
		defer of.Close()
		w = of
	}

	t := template.NewTemplate(inPath, f.String(), w)
	if err := t.Render(); err != nil {
		return 1, err
	}

	return 0, nil
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"template": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Render a Go template using environment variables."
}
