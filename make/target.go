package make

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkar/runit"
	yaml "gopkg.in/yaml.v2"

	"github.com/upsight/ron/color"
	"github.com/upsight/ron/execute"
)

// Target contains the set of commands to run along with
// any before and after targets to run.
type Target struct {
	Name         string        `json:"name" yaml:"name"`
	Before       []string      `json:"before" yaml:"before"`
	After        []string      `json:"after" yaml:"after"`
	Cmd          string        `json:"cmd" yaml:"cmd"`
	Description  string        `json:"description" yaml:"description"`
	IsDefault    bool          `json:"isDefault" yaml:"isDefault"`
	TargetConfig *TargetConfig `json:"-" yaml:"-"`
	W            io.Writer     `json:"-" yaml:"-"` // underlying stdout writer
	WErr         io.Writer     `json:"-" yaml:"-"` // underlying stderr writer
}

// TargetConfig is a mapping of target names to target
// commands.
type TargetConfig struct {
	Targets map[string]*Target
	Env     *Env // the environment variables used for commands.
	StdOut  io.Writer
	StdErr  io.Writer
	configs []*Config
}

// NewTargetConfig takes a default set of yaml in config format and then
// overrides them with a new set of config target replacements.
func NewTargetConfig(env *Env, configs []*Config, stdOut io.Writer, stdErr io.Writer) (*TargetConfig, error) {
	if stdOut == nil {
		stdOut = os.Stdout
	}
	if stdErr == nil {
		stdErr = os.Stderr
	}

	t := &TargetConfig{
		Env:     env,
		Targets: map[string]*Target{},
		StdOut:  stdOut,
		StdErr:  stdErr,
		configs: configs,
	}

	for _, config := range t.configs {
		var targets map[string]*Target
		if err := yaml.Unmarshal([]byte(config.Targets), &targets); err != nil {
			return nil, err
		}
		for name, target := range targets {
			t.Targets[name] = target
			t.Targets[name].IsDefault = config.IsDefault
		}
	}
	// initialize io for each target.
	for name, target := range t.Targets {
		target.W = stdOut
		target.WErr = stdErr
		target.TargetConfig = t
		target.Name = name
	}
	return t, nil
}

// List prints out each target and its before and after targets.
func (tc *TargetConfig) List(verbose bool, fuzzy string) {
	targetNameWidth := 0
	targetNames := []string{}
	targetNamesDefault := []string{}
	for k, target := range tc.Targets {
		if len(k) > targetNameWidth {
			targetNameWidth = len(k)
		}
		if fuzzy != "" {
			if ok, _ := filepath.Match(fuzzy, k); !ok {
				continue
			}
		}
		switch target.IsDefault {
		case false:
			targetNames = append(targetNames, k)
		default:
			targetNamesDefault = append(targetNamesDefault, k)
		}
	}
	sort.Strings(targetNames)
	sort.Strings(targetNamesDefault)
	tc.StdOut.Write([]byte(color.Green("Targets:\n\n")))
	for _, targetName := range targetNames {
		if target, ok := tc.Target(targetName); ok {
			target.List(verbose, targetNameWidth)
		}
	}
	tc.StdOut.Write([]byte(color.Green("\nDefault Targets:\n\n")))
	for _, targetName := range targetNamesDefault {
		if target, ok := tc.Target(targetName); ok {
			target.List(verbose, targetNameWidth)
		}
	}
}

// Target retrieves the named target from config. If it doesn't
// exists a bool false will be returned along with nil
func (tc *TargetConfig) Target(name string) (*Target, bool) {
	target, ok := tc.Targets[name]
	return target, ok
}

// runTargetList executes a list of targets.
func (t *Target) runTargetList(targets []string) (int, string, error) {
	for _, target := range targets {
		if target == t.Name {
			continue
		}
		if t, ok := t.TargetConfig.Target(target); ok {
			status, out, err := t.Run()
			if status != 0 || err != nil {
				return status, out, err
			}
		}
	}
	return 0, "", nil
}

// Run executes the targets before commands then runs its own
// followed by after targets. Try not to make circular references please.
func (t *Target) Run() (int, string, error) {
	if len(t.Before) > 0 {
		status, out, err := t.runTargetList(t.Before)
		if status != 0 || err != nil {
			return status, out, err
		}
	}

	cmd, err := execute.CommandNoWait(t.Cmd, t.W, t.WErr, t.TargetConfig.Env.Config)
	if err != nil {
		return 1, "", err
	}
	interrupt := make(chan os.Signal, 1)
	go func(c *exec.Cmd) {
		execute.WaitNoop(interrupt, cmd)
		if c != nil && c.Process != nil {
			c.Process.Kill()
		}
	}(cmd)
	err = cmd.Wait()
	if err != nil {
		status := runit.GetExitStatus(err)
		return status, "", err
	}

	if len(t.After) > 0 {
		status, out, err := t.runTargetList(t.After)
		if status != 0 || err != nil {
			return status, out, err
		}
	}
	return 0, "", nil
}

// List displays the defined before, after, description and cmd of the target.
func (t *Target) List(verbose bool, nameWidth int) {
	if !verbose {
		if strings.HasPrefix(t.Name, "_") {
			// skip targets in non verbose mode as hidden.
			return
		}
		padWidth := nameWidth - len(t.Name)
		paddedName := color.Yellow(t.Name)
		if padWidth > 0 {
			paddedName += strings.Repeat(" ", padWidth)
		}
		out := fmt.Sprintf("%s %s\n", paddedName, strings.TrimSpace(t.Description))
		_, err := t.W.Write([]byte(out))
		if err != nil {
			log.Println(color.Red(err.Error()))
		}
		return
	}

	// target name
	out := fmt.Sprintf("%s: \n", color.Yellow(t.Name))

	// target description
	if t.Description != "" {
		out += fmt.Sprintf("  - description: %s\n", strings.TrimSpace(t.Description))
	}

	// target before
	if len(t.Before) > 0 {
		beforeList := "  - before: " + strings.Join(t.Before, ", ")
		out += fmt.Sprintln(beforeList)
	}

	// target after
	if len(t.After) > 0 {
		afterList := "  - after: " + strings.Join(t.After, ", ")
		out += fmt.Sprintln(afterList)
	}

	// target command
	out += fmt.Sprintf("  - cmd:\n    ")
	out += fmt.Sprintln(strings.Replace(t.Cmd, "\n", "\n    ", -1))
	_, err := t.W.Write([]byte(out))
	if err != nil {
		log.Println(color.Red(err.Error()))
	}
}
