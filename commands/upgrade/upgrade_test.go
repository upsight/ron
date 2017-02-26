package ron

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/upsight/ron/target"
)

var (
	latestBinaryServer   *httptest.Server
	mockLoadLatestBinary = func(url, tmpPath, path string) error {
		return nil
	}
	mockLoadLatestBinaryErr = func(url, tmpPath, path string) error {
		return fmt.Errorf("err")
	}
)

func init() {
	latestBinaryServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/notfound") {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, []byte("binaryfile"))
	}))
}

func TestRonRunUpgrade(t *testing.T) {
	/*
		TODO
			prevLoadLatestBinary := loadLatestBinary
			defer func() { loadLatestBinary = prevLoadLatestBinary }()
			loadLatestBinary = mockLoadLatestBinary

			args := []string{}
			stdOut := &bytes.Buffer{}
			stdErr := &bytes.Buffer{}
			c := &Command{W: stdOut, WErr: stdErr, AppName: "a"}
			c.Run(args)
			if stdErr.String() != "" {
				t.Fatalf("expected empty err got %s", stdErr.String())
			}
			want := fmt.Sprintf("%s\n", c.AppName)
			if stdOut.String() != want {
				t.Logf("upgrade command want %s got %s", want, stdOut.String())
			}
	*/
}

func TestRonRunUpgradeBadEnvs(t *testing.T) {
	prevEnvs := target.DefaultEnvConfig
	defer func() { target.DefaultEnvConfig = prevEnvs }()
	target.DefaultEnvConfig = `envs:"`
	args := []string{}
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr, AppName: "a"}
	status, err := c.Run(args)
	if status == 0 {
		t.Errorf("expected status non 0 got %d %+v", status, err)
	}
}

func TestRonRunUpgradeMissingLatestURL(t *testing.T) {
	prevEnvs := target.DefaultEnvConfig
	defer func() { target.DefaultEnvConfig = prevEnvs }()
	target.DefaultEnvConfig = `envs:`
	args := []string{}
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr, AppName: "a"}
	status, err := c.Run(args)
	if status != 1 && err.Error() != "LATEST_URL env key not set" {
		t.Errorf("expected status 1 got %d %+v", status, err)
	}
}

func TestRonRunUpgradeLoadLatestBinaryErr(t *testing.T) {
	prevLoadLatestBinary := loadLatestBinary
	defer func() { loadLatestBinary = prevLoadLatestBinary }()
	loadLatestBinary = mockLoadLatestBinaryErr

	args := []string{}
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c := &Command{W: stdOut, WErr: stdErr, AppName: "a"}
	status, err := c.Run(args)
	if status != 1 && err.Error() != "err" {
		t.Errorf("expected status 1 got %d %+v", status, err)
	}
}

func TestRonUpgradeLoadLatestBinary(t *testing.T) {
	AppName := "a"
	prevAppName := AppName
	defer func() { AppName = prevAppName }()
	AppName = "Scotchy"

	binPath := filepath.Join(os.TempDir(), AppName)
	err := loadLatestBinary(latestBinaryServer.URL, os.TempDir(), binPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(binPath)

	if _, err := os.Stat(binPath); err != nil {
		t.Fatal("binPath not written with new binary")
	}
}

func TestRonRunUpgradeLoadLatestBinaryBadTmpFile(t *testing.T) {
	err := loadLatestBinary(latestBinaryServer.URL, "/dev/null", "")
	if err == nil {
		t.Fatal("expected err for not being able to write tmp file")
	}
}

func TestRonUpgradeLoadLatestBinaryOpenFail(t *testing.T) {
	err := loadLatestBinary(latestBinaryServer.URL, os.TempDir(), "")
	if err == nil {
		t.Fatal("expected err for open file")
	}
}

func TestRonUpgradeLoadLatestBinaryNotFound(t *testing.T) {
	err := loadLatestBinary(latestBinaryServer.URL+"/notfound", os.TempDir(), "")
	if err == nil {
		t.Fatal("expected err for not found url")
	}
}
