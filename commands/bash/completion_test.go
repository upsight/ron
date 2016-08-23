package bash

import (
	"bytes"
	"testing"
)

func TestRonBashCompletion(t *testing.T) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr}
	c.Run([]string{})
	wantErr := `Copy the following into your ~/.bashrc file or into /etc/bash_completion/ron`
	want := ronComplete
	if stdErr.String() != wantErr {
		t.Fatalf("bash_completion command want %s got %s", wantErr, stdErr.String())
	}

	if stdOut.String() != want {
		t.Fatalf("bash_completion command want %s got %s", want, stdOut.String())
	}
}
