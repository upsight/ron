package execute

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/upsight/ron/color"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSH holds the ssh configuration and io.
type SSH struct {
	Config *SSHConfig
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// SSHConfig is an individual ssh configuration for creating
// a connection.
type SSHConfig struct {
	Host         string `json:"host" yaml:"host"`
	Port         int    `json:"port" yaml:"port"`
	User         string `json:"user" yaml:"user"`
	ProxyHost    string `json:"proxy_host,omitempty" yaml:"proxy_host,omitempty"`
	ProxyPort    int    `json:"proxy_port,omitempty" yaml:"proxy_port,omitempty"`
	ProxyUser    string `json:"proxy_user,omitempty" yaml:"proxy_user,omitempty"`
	IdentityFile string `json:"identity_file,omitempty" yaml:"identity_file,omitempty"`
}

// RunCommand will execute a command using the input environment variables.
// It will establish a new session on every call.
func (s *SSH) RunCommand(cmd string, envs map[string]string) error {
	var (
		conn       *ssh.Client
		err        error
		authMethod ssh.AuthMethod
	)
	switch {
	case s.Config.IdentityFile != "":
		pemBytes, err := ioutil.ReadFile(s.Config.IdentityFile)
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return err
		}
		authMethod = ssh.PublicKeys(signer)
	default:
		sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err != nil {
			return err
		}
		authMethod = ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	targetAddr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)
	targetConfig := &ssh.ClientConfig{
		User: s.Config.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	switch {
	case s.Config.ProxyHost != "":
		bastionAddr := fmt.Sprintf("%s:%d", s.Config.ProxyHost, s.Config.ProxyPort)
		bastionConfig := &ssh.ClientConfig{
			User: s.Config.ProxyUser,
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		// connect to the bastion host
		bastionClient, err := ssh.Dial("tcp", bastionAddr, bastionConfig)
		if err != nil {
			return fmt.Errorf("unable to connect: %s", err)
		}
		defer bastionClient.Close()

		// dial a connection to the service host, from the bastion
		bconn, err := bastionClient.Dial("tcp", targetAddr)
		if err != nil {
			return fmt.Errorf("unable to connect: %s", err)
		}
		defer bconn.Close()

		ncc, chans, reqs, err := ssh.NewClientConn(bconn, targetAddr, targetConfig)
		if err != nil {
			return fmt.Errorf("failed to create conn: %s", err)
		}
		conn = ssh.NewClient(ncc, chans, reqs)
	default:
		conn, err = ssh.Dial("tcp", targetAddr, targetConfig)
		if err != nil {
			return fmt.Errorf("unable to connect: %s", err)
		}
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
				fmt.Fprintf(s.Stdout, color.Green("%s]")+" %s\n", s.Config.Host, scanner.Text())
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
				fmt.Fprintf(s.Stderr, color.Red("%s]")+" %s\n", s.Config.Host, scanner.Text())
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
func NewSSH(conf *SSHConfig, stdin io.Reader, stdout, stderr io.Writer) (*SSH, error) {
	s := &SSH{
		Config: conf,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	return s, nil
}
