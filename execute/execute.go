// Package execute tries it's best to run complex, sometimes multiline bash commands.
package execute

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/upsight/ron/color"
)

var (
	// Debug prints the command being run if set to true.
	Debug = false
)

// WaitNoop accepts and waits on any signal and returns on kill signals.
func WaitNoop(interrupt chan os.Signal, cmd *exec.Cmd) {
	signal.Notify(interrupt)
	for {
		select {
		case sig := <-interrupt:
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
				return
			default:
				continue
			}
		}
	}
}

func getCmd(cmdString string, stdOut io.Writer, stdErr io.Writer, envs map[string]string) *exec.Cmd {
	if Debug {
		switch {
		case envs != nil:
			// os.Expand doesn't work well with $( so replace it with
			// something that won't alter and revert.
			c := strings.Replace(cmdString, "$(", "Ω(", -1)
			getEnvFunc := func(k string) string {
				v, _ := envs[k]
				return v
			}
			c = os.Expand(c, getEnvFunc)
			c = strings.Replace(c, "Ω(", "$(", -1)
			c = strings.Replace(c, "\n", "\n\t", -1)
			fmt.Println(color.Blue("\t" + c))
		default:
			fmt.Println(color.Blue(os.ExpandEnv(cmdString)))
		}
	}
	cmd := exec.Command("bash", "-e", "-c", cmdString)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	if envs != nil {
		cmd.Env = []string{}
		for k, v := range envs {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return cmd
}

// Command just executes a given cmd string to the supplied io.Writer writers.
// If optional envs is passed in then the expanded values will be used vs the os versions.
func Command(cmdString string, stdOut io.Writer, stdErr io.Writer, envs map[string]string) (int, error) {
	cmd := getCmd(cmdString, stdOut, stdErr, envs)
	exitStatus := 0
	err := cmd.Run()
	if err != nil {
		exitStatus = GetExitStatus(err)
	}
	return exitStatus, err
}

// CommandNoWait starts the given command but does not wait for it to finish. It returns
// the created exec.Command which can be used with Wait.
func CommandNoWait(cmdString string, stdOut io.Writer, stdErr io.Writer, envs map[string]string) (*exec.Cmd, error) {
	cmd := getCmd(cmdString, stdOut, stdErr, envs)
	return cmd, cmd.Start()
}

// GetExitStatus determines the exit status code of an err
// from a command that was run.
func GetExitStatus(waitError error) int {
	if exitError, ok := waitError.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return waitStatus.ExitStatus()
		}
	}
	return 1
}
