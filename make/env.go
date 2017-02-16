package make

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/upsight/ron/color"
	"github.com/upsight/ron/execute"
)

const (
	// SENTINEL is the first character looked for in envs to
	// signify that the value should be executed in the shell
	// and the output assigned to the key.
	SENTINEL = "+"
)

// Env takes a raw yaml environment definition and expands and
// overrides any variables.
type Env struct {
	Config         MSS       // the key value of expanded variables
	W              io.Writer // underlying writer
	OSEnvs         MSS       // the initial environment variables
	rawDefaultEnvs string    // the raw default env configuration yaml
	rawNewEnvs     string    // the raw overriden env configuration yaml
	keyOrder       []string  // the env keys order of preference
}

// ParseOSEnvs takes a list of "key=val" and splits them
// into a map[string]string.
func ParseOSEnvs(osEnvs []string) MSS {
	envs := MSS{}
	for _, env := range osEnvs {
		pair := strings.SplitN(env, "=", 2)
		envs[pair[0]] = pair[1]
	}
	return envs
}

// NewEnv create a new environment variable parser similar
// to make variables.
func NewEnv(rawDefaultEnvs, rawNewEnvs string, osEnvs MSS, writer io.Writer) (*Env, error) {
	if writer == nil {
		writer = os.Stdout
	}

	e := &Env{
		W:              writer,
		Config:         MSS{},
		OSEnvs:         osEnvs,
		rawDefaultEnvs: rawDefaultEnvs,
		rawNewEnvs:     rawNewEnvs,
		keyOrder:       []string{},
	}
	err := e.process()
	if err != nil {
		return nil, err
	}

	return e, nil
}

func in(key string, keys []string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// process takes the raw env configuration yaml and converts
// it to expanded variable definitions based on passed in
// environment variables and yaml config.
// The overriding value used is from os.Environ.
func (e *Env) process() error {
	// Load default configs
	var defaultEnvs []MSS
	err := yaml.Unmarshal([]byte(e.rawDefaultEnvs), &defaultEnvs)
	if err != nil {
		return err
	}
	// Load any overriden configs
	var rawNewEnvs []MSS
	err = yaml.Unmarshal([]byte(e.rawNewEnvs), &rawNewEnvs)
	if err != nil {
		return err
	}

	// load default envs
	for _, env := range defaultEnvs {
		for k := range env {
			e.Config[k] = env[k]
			if !in(k, e.keyOrder) {
				e.keyOrder = append(e.keyOrder, k)
			}
		}
	}
	// load any new envs
	for _, env := range rawNewEnvs {
		for k := range env {
			e.Config[k] = env[k]
			// Append new envs to the key order list
			// so that this order takes precedence.
			if !in(k, e.keyOrder) {
				e.keyOrder = append(e.keyOrder, k)
			}
		}
	}
	// Load any os enviroment as final value
	for k, v := range e.OSEnvs {
		e.Config[k] = v
	}

	// env variable values that start with SENTINEL will be
	// executed to get the output value and set.
	// All variables expand any envs defined in order of definition.
	for _, k := range e.keyOrder {
		if strings.HasPrefix(e.Config[k], SENTINEL) {
			out, err := e.getExec(k)
			if err != nil {
				return err
			}
			e.Config[k] = out
		}
		e.Config[k] = os.Expand(e.Config[k], e.Getenv)
	}
	return nil
}

// getExec executes the command defined by the SENTINEL string.
func (e *Env) getExec(key string) (out string, err error) {
	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}
	status, err := execute.Command(e.Config[key][1:], &stdOut, &stdErr, e.Config)
	switch {
	case status == 0:
		out = strings.TrimSpace(stdOut.String())
	case err != nil:
		err = fmt.Errorf("stdout: %s strderr: %s", stdOut.String(), stdErr.String())
	default:
		err = fmt.Errorf("status code: %d", status)
	}
	return
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
func (e *Env) Getenv(key string) string {
	v, _ := e.Config[key]
	if strings.HasPrefix(v, SENTINEL) {
		out, err := e.getExec(key)
		if err == nil {
			v = out
		}
	}
	return v
}

// List prints to the underlying writer a list of
// the configured env based on overriden environment
// variables and default yaml ones.
func (e *Env) List() error {
	envNameWidth := 0
	for _, k := range e.keyOrder {
		if len(k) > envNameWidth {
			envNameWidth = len(k)
		}
	}
	for _, k := range e.keyOrder {
		padWidth := envNameWidth - len(k)
		paddedKey := ""
		if padWidth > 0 {
			paddedKey += strings.Repeat(" ", padWidth)
			paddedKey += k
		}
		_, err := e.W.Write([]byte(fmt.Sprintln(color.Green(paddedKey+"=") + e.Config[k])))
		if err != nil {
			return err
		}
	}
	return nil
}

// PrintRaw outputs the unprocessed yaml given to Env in both
// the defaults and overriden.
func (e *Env) PrintRaw() error {
	_, err := e.W.Write([]byte(e.rawDefaultEnvs + "\n"))
	if err != nil {
		return err
	}
	_, err = e.W.Write([]byte(e.rawNewEnvs + "\n"))
	if err != nil {
		return err
	}
	return nil
}
