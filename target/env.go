package target

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
	// ExecSentinel is the first character looked for in envs to
	// signify that the value should be executed in the shell
	// and the output assigned to the key.
	ExecSentinel = "+"
)

// Env takes a raw yaml environment definition and expands and
// overrides any variables.
type Env struct {
	OSEnvs      MSS       // the initial environment variables
	W           io.Writer // underlying writer
	config      MSS       // the key value of expanded variables
	keyOrder    []string  // the env keys order of preference
	rawConfig   *RawConfig
	parent      *File
	isProcessed bool
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
// to make variables. In order to populate the Config envs, the func Process
// must be run after initialization.
func NewEnv(parentFile *File, config *RawConfig, osEnvs MSS, writer io.Writer) (*Env, error) {
	if writer == nil {
		writer = os.Stdout
	}

	e := &Env{
		OSEnvs:    osEnvs,
		W:         writer,
		config:    MSS{},
		keyOrder:  []string{},
		rawConfig: config,
		parent:    parentFile,
	}
	err := e.initEnvKeyOrder()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Config returns the envs config as a map[string]string. It will process
// each env if that has not been done already.
func (e *Env) Config() (MSS, error) {
	if !e.isProcessed {
		err := e.process()
		if err != nil {
			return nil, err
		}
	}
	return e.config, nil
}

// MergeTo the current env into any missing keys for the input node.
func (e *Env) MergeTo(node *Env) error {
	for k, v := range e.config {
		if _, ok := node.config[k]; !ok {
			node.config[k] = v
			if !keyIn(k, node.keyOrder) {
				node.keyOrder = append(node.keyOrder, k)
			}
		}
	}
	return nil
}

// initEnvKeyOrder initialize the internal config mapping and key order.
func (e *Env) initEnvKeyOrder() error {
	var envs []MSS
	if err := yaml.Unmarshal([]byte(e.rawConfig.Envs), &envs); err != nil {
		return err
	}
	for _, env := range envs {
		for k, v := range env {
			e.config[k] = v
			if e.parent != nil {
				// use the parents env here if it exists
				if eParent, ok := e.parent.Env.config[k]; ok {
					e.config[k] = eParent
				}
			}
			if !keyIn(k, e.keyOrder) {
				e.keyOrder = append(e.keyOrder, k)
			}
		}
	}
	return nil
}

// process takes the raw env configuration yaml and converts
// it to expanded variable definitions based on passed in
// environment variables and yaml config.
// The overriding value used is from os.Environ.
func (e *Env) process() error {
	if e.isProcessed {
		// already processed
		return nil
	}
	if e.parent != nil {
		err := e.parent.Env.process()
		if err != nil {
			return err
		}
	}
	for k, v := range e.OSEnvs {
		e.config[k] = v
	}
	e.isProcessed = true

	// env variable values that start with ExecSentinel will be
	// executed to get the output value and set.
	// All variables expand any envs defined in order of definition.
	for _, k := range e.keyOrder {
		if strings.HasPrefix(e.config[k], ExecSentinel) {
			out, err := e.getExec(k)
			if err != nil {
				return err
			}
			e.config[k] = out
		}
		e.config[k] = os.Expand(e.config[k], e.Getenv)
	}
	if e.parent != nil {
		// set the final envs from parent here as final
		// only if the value is empty.
		for k, v := range e.parent.Env.config {
			if e.config[k] == "" {
				e.config[k] = v
			}
		}
	}
	return nil
}

// getExec executes the command defined by the ExecSentinel string.
func (e *Env) getExec(key string) (out string, err error) {
	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}
	status, err := execute.Command(e.config[key][1:], &stdOut, &stdErr, e.config)
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
	if !e.isProcessed {
		err := e.process()
		if err != nil {
			return ""
		}
	}
	v, _ := e.config[key]
	if strings.HasPrefix(v, ExecSentinel) {
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
	if !e.isProcessed {
		err := e.process()
		if err != nil {
			return err
		}
	}
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
		} else {
			paddedKey = k
		}
		_, err := e.W.Write([]byte(fmt.Sprintln(color.Green(paddedKey+"=") + e.config[k])))
		if err != nil {
			return err
		}
	}
	return nil
}

// PrintRaw outputs the unprocessed yaml given to Env in both
// the defaults and overriden.
func (e *Env) PrintRaw() error {
	_, err := e.W.Write([]byte(e.rawConfig.Envs + "\n"))
	if err != nil {
		return err
	}
	return nil
}
