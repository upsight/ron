package target

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/upsight/ron/color"
	"github.com/upsight/ron/execute"
)

// Target contains the set of commands to run along with
// any before and after targets to run.
type Target struct {
	targetConfigs *Configs
	File          *File     `json:"-" yaml:"-"`
	Name          string    `json:"name" yaml:"name"`
	Before        []string  `json:"before" yaml:"before"`
	After         []string  `json:"after" yaml:"after"`
	Cmd           string    `json:"cmd" yaml:"cmd"`
	Description   string    `json:"description" yaml:"description"`
	W             io.Writer `json:"-" yaml:"-"` // underlying stdout writer
	WErr          io.Writer `json:"-" yaml:"-"` // underlying stderr writer
}

// runTargetList executes a list of targets.
func (t *Target) runTargetList(targets []string) (int, string, error) {
	for _, target := range targets {
		if target == t.Name {
			continue
		}
		if t, ok := t.targetConfigs.Target(target); ok {
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

	envs, err := t.File.Env.Config()
	if err != nil {
		return 1, "", err
	}
	cmd, err := execute.CommandNoWait(t.Cmd, t.W, t.WErr, envs)
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
		status := execute.GetExitStatus(err)
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

// RunRemote executes the target on a remote host. It ignores any
// before and after targets.
func (t *Target) RunRemote(conf *execute.SSHConfig) (int, string, error) {
	s, err := execute.NewSSH(conf, os.Stdin, t.W, t.WErr)
	if err != nil {
		return 1, "", err
	}

	err = s.RunCommand(t.Cmd, nil)
	return 0, "", err
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
