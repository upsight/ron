package cmd

import (
	"bytes"
	"testing"
)

func TestRonCmdRunCmd(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, err := c.Run([]string{"ls"})
	if status != 0 {
		t.Fatalf("status not 0 got %d %s", status, stdOut.String())
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestRonCmdRunCmdNArgsErr(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"ls nope"})
	if status < 1 {
		t.Fatalf("status not > 0 got %d", status)
	}
}

func TestRonCmdRunnerErr(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"ls nothinghere"})
	if status == 0 {
		t.Fatalf("status is not non 0 got %v", status)
	}
}

func TestRonCmdArgsExitErr(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"exit 1"})
	if status != 1 {
		t.Fatalf("status is not 1 got %v", status)
	}
}

func TestRonCmdArgsParseErr(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, _ := c.Run([]string{"-"})
	if status != 1 {
		t.Fatalf("status is not 1 got %v", status)
	}
}

func TestRonCmdRunitErr(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	status, err := c.Run([]string{"--watch=nothere", "ls"})
	if status != 1 {
		t.Fatalf("status is not 1 got %v", status)
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRonCmdLongRunning(t *testing.T) {
	/*
		TODO
		cmd, _, _ := createCommand("cmd")
		cmd.Interrupt = make(chan os.Signal, 1)
		var status int
		var err error
		go func(s int, e error) {
			status, err = runCmd(cmd, []string{"--restart", "ls"})
		}(status, err)
		cmd.Interrupt <- syscall.SIGINT
		if status != 0 || err != nil {
			t.Errorf("expected status 0 err nil got %d %s", status, err)
		}
	*/
}
