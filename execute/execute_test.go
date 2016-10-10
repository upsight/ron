package execute

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestExecuteCommand(t *testing.T) {
	os.Setenv("__TESTRON__", "a-b")
	defer func() { os.Unsetenv("__TESTRON__") }()

	type outTest struct {
		status    int
		output    string
		errString string
		err       error
	}
	type tableTest struct {
		in   string
		envs map[string]string
		out  outTest
	}
	tableTests := []tableTest{
		{`echo "prep1"
			echo "prep2" && \
				echo "prep3"
			if [ 1 != 0 ]; then \
				echo "prep4"
			else \
				echo "prepnope"; \
			fi
			`, nil, outTest{0, "prep1\nprep2\nprep3\nprep4\n", "", nil}},
		{`echo "prep1"
			echo "prep2" && echo "prep3"
			if [ 1 != 1 ]; then
				echo "prep4"
			else
				echo "prepnope";
			fi
			`, nil, outTest{0, "prep1\nprep2\nprep3\nprepnope\n", "", nil}},
		{`echo $(echo "prep1")`, nil, outTest{0, "prep1\n", "", nil}},
		{`exit 1`, nil, outTest{1, "", "", nil}},
		{`exit 2`, nil, outTest{2, "", "", nil}},
		{`echo $__TESTRON__ | tr "-" "_"`, nil, outTest{0, "a_b\n", "", nil}},
		{`_vermouth`, nil, outTest{127, "", fmt.Sprintf("bash: _vermouth: command not found\n"), nil}},
		{`echo $HI $RON`, map[string]string{"RON": "ron", "HI": "hi"}, outTest{0, "hi ron\n", "", nil}},
	}
	for i, test := range tableTests {
		var outBuf bytes.Buffer
		var errBuf bytes.Buffer
		status, _ := Command(test.in, &outBuf, &errBuf, test.envs)
		if status != test.out.status {
			t.Errorf(`%d status want "%+v" got "%+v"`, i, test.out.status, status)
		}
		if test.out.errString != errBuf.String() {
			t.Errorf(`%d error want "%v", got "%v"`, i, test.out.errString, errBuf.String())
		}
		if outBuf.String() != test.out.output {
			t.Errorf(`%d input: "%s" want "%+v" got "%+v"`, i, test.in, test.out.output, outBuf.String())
		}
	}
}

func TestExecuteCommandIO(t *testing.T) {
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	Command("echo prep", &outBuf, &errBuf, nil)
	if !strings.Contains(outBuf.String(), "prep") {
		t.Errorf("output mismatch got %s want prep", outBuf.String())
	}
	if errBuf.String() != "" {
		t.Fatal(errBuf.String())
	}
}

func TestExecuteCommandDEBUG(t *testing.T) {
	Debug = true
	defer func() { Debug = false }()
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	Command("echo prep", &outBuf, &errBuf, nil)
}

func TestExecuteCommandNoWait(t *testing.T) {
	Debug = true
	defer func() { Debug = false }()

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd, err := CommandNoWait("while true; do sleep 1; done", &outBuf, &errBuf, nil)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		cmd.Process.Kill()
	}()
	err = cmd.Wait()
	if err == nil {
		t.Error("process ")
	}
}

func TestExecuteWaitNoop(t *testing.T) {
	interrupt := make(chan os.Signal, 1)
	go func(i chan os.Signal) {
		WaitNoop(i, &exec.Cmd{})
	}(interrupt)
	interrupt <- syscall.SIGHUP
	time.Sleep(100 * time.Millisecond)
}

func TestExecuteWaitNoopKill(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	interrupt := make(chan os.Signal, 1)
	go func(i chan os.Signal) {
		WaitNoop(i, cmd)
	}(interrupt)
	interrupt <- syscall.SIGINT
	cmd.Wait()
}
