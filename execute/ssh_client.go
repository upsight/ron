package execute

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/upsight/ron/color"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSH holds the ssh configuration and io.
type SSH struct {
	Config *ssh.ClientConfig
	Host   string
	Port   int
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// RunCommand will execute a command using the input environment variables.
// It will establish a new session on every call.
func (s *SSH) RunCommand(cmd string, envs map[string]string) error {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port), s.Config)
	if err != nil {
		return fmt.Errorf("unable to connect: %s", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %s", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	err = s.prepareCommand(session, cmd, envs)
	if err != nil {
		return err
	}

	err = session.Start(cmd)
	if err != nil {
		return err
	}
	return session.Wait()
}

func (s *SSH) prepareCommand(session *ssh.Session, cmd string, envs map[string]string) error {
	for k, v := range envs {
		if err := session.Setenv(k, v); err != nil {
			return err
		}
	}

	if s.Stdin != nil {
		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("unable to setup stdin for session: %v", err)
		}
		go io.Copy(stdin, s.Stdin)
	}

	if s.Stdout != nil {
		stdout, err := session.StdoutPipe()
		if err != nil {
			return fmt.Errorf("unable to setup stdout for session: %v", err)
		}
		scanner := bufio.NewScanner(stdout)
		go func() {
			for scanner.Scan() {
				fmt.Fprintf(s.Stdout, color.Green("%s]")+" %s\n", s.Host, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(s.Stderr, err)
			}
		}()
	}

	if s.Stderr != nil {
		stderr, err := session.StderrPipe()
		if err != nil {
			return fmt.Errorf("unable to setup stderr for session: %v", err)
		}
		scanner := bufio.NewScanner(stderr)
		go func() {
			for scanner.Scan() {
				fmt.Fprintf(s.Stderr, color.Red("%s]")+" %s\n", s.Host, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(s.Stderr, err)
			}
		}()
	}
	return nil
}

// NewSSH will initialize a new ssh configuration for use with RunCommand.
// This assumes an ssh agent is available and the proper keys have been added
// with ssh-add. You can check added keys with ssh-add -L
func NewSSH(user, host string, port int, stdin io.Reader, stdout, stderr io.Writer) (*SSH, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
		},
		Timeout: time.Duration(5 * time.Second),
	}

	s := &SSH{
		Config: sshConfig,
		Host:   host,
		Port:   port,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	return s, nil
}
