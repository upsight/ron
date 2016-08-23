package template

import (
	"os"
	"strings"
)

// Context represents the environment variables to be used in
// rendering templates.
type Context struct {
	env map[string]string
}

// Env returns the os environment variables.
func (c *Context) Env() map[string]string {
	return c.env
}

// NewContext initializes from environment variables.
func NewContext() *Context {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return &Context{env}
}
