package burgundy

import (
	"bytes"
	"strings"
	"testing"
)

func TestRonRunBurgundy(t *testing.T) {
	stdOut := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: nil}
	c.Run([]string{})

	if !strings.Contains(stdOut.String(), "░▒▒╣▒▓▓▓▓▓█▒") {
		t.Fatal("ron isn't there")
	}
}
