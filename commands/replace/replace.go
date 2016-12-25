package replace

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	// RonTmpExt is the extension for backup files.
	RonTmpExt = ".ron.tmp"
)

// Command ...
type Command struct {
	W       io.Writer
	WErr    io.Writer
	AppName string
	Name    string
	Debug   bool
}

func shouldIgnore(path string) bool {
	switch {
	case strings.Contains(path, ".git/"):
		return true
	case strings.Contains(path, ".hg/"):
		return true
	case strings.HasSuffix(path, RonTmpExt):
		return true
	}
	return false
}

func isTextFile(b []byte, fi os.FileInfo) bool {
	return utf8.Valid(b)
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	fs := flag.NewFlagSet(c.Name, flag.ExitOnError)
	fs.SetOutput(c.WErr)
	fs.Usage = func() {
		fmt.Fprintf(c.W, "Usage: %s %s [-debug] [path] [replace] [replacewith]\n", c.AppName, c.Name)
		fs.PrintDefaults()
	}
	fs.BoolVar(&c.Debug, "debug", false, "Debug the replace run")
	fs.Parse(args)
	if len(args) < 3 {
		fs.Usage()
		return 1, nil
	}
	path := fs.Args()[0]
	rpl := fs.Args()[1]
	rplWith := fs.Args()[2]

	filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			if c.Debug {
				c.WErr.Write([]byte(fmt.Sprintln("ERRO: ", err.Error())))
			}
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		if shouldIgnore(path) {
			return nil
		}
		if fi.Size() > 10485760 {
			if c.Debug {
				c.WErr.Write([]byte(fmt.Sprintln("ERRO: filesize larger than 10mb", path, fi.Size())))
			}
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			if c.Debug {
				c.WErr.Write([]byte(fmt.Sprintln("ERRO: ", err.Error())))
			}
			return nil
		}
		b := make([]byte, 512)
		_, err = f.Read(b)
		f.Close()
		if err != nil {
			if c.Debug {
				c.WErr.Write([]byte(fmt.Sprintln("ERRO: ", err.Error())))
			}
			return nil
		}

		if isTextFile(b, fi) {
			input, err := ioutil.ReadFile(path)
			if err != nil {
				if c.Debug {
					c.WErr.Write([]byte(fmt.Sprintln("ERRO: ", err.Error())))
				}
				return nil
			}
			count := strings.Count(string(input), rpl)
			if count > 0 {
				output := bytes.Replace(input, []byte(rpl), []byte(rplWith), -1)
				err = ioutil.WriteFile(path+RonTmpExt, output, fi.Mode())
				if err != nil {
					c.WErr.Write([]byte(fmt.Sprintln("ERRO: ", err.Error())))
					return nil
				}
				err = os.Rename(path+RonTmpExt, path)
				if err != nil {
					c.WErr.Write([]byte(fmt.Sprintln("ERRO: ", err.Error())))
					return nil
				}
				c.W.Write([]byte(fmt.Sprintf("REPLACED %d: %s\n", count, path)))
			}
		}
		return nil
	})

	return 0, nil
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"replace": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Find and replace in text."
}
