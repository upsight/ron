package upgrade

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/upsight/ron/execute"
	"github.com/upsight/ron/target"
)

// Command ...
type Command struct {
	Name    string
	W       io.Writer
	WErr    io.Writer
	AppName string
}

// loadLatestBinary downloads the latest ron from the configured url
// and overwrites the currently running binary with it.
var loadLatestBinary = func(binUrl, binTmpDir, binPath string) error {
	// create tmp file to download to
	tmpFilePath := binPath + ".tmp"
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	// download the file and copy to the tmp file.
	log.Println("downloading ron from LATEST_URL=", binUrl)
	resp, err := http.Get(binUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("http error: %d %s", resp.StatusCode, msg)
	}
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}
	log.Println("ron downloaded to", tmpFilePath)

	err = os.Rename(tmpFilePath, binPath)
	if err != nil {
		return err
	}
	err = os.Chmod(binPath, 0777)
	if err != nil {
		return err
	}

	log.Println("ron installed")
	return nil
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// findPath calculates the path to the currently running executable.
func (c *Command) findPath() (string, error) {
	// get the path to the current version installed
	binPath, err := exec.LookPath(c.AppName)
	if err != nil {
		// Try just getting it from the current arg 0
		// This only works if the executable is run as ./ron upgrade
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return "", err
		}
		binPath = path.Join(dir, filepath.Base(os.Args[0]))
	}

	return binPath, nil
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	// get the path to the current version installed
	binPath, err := c.findPath()
	if err != nil {
		return 1, err
	}
	// Create envs
	envsConfig, _, err := target.BuiltinDefault()
	if err != nil {
		return 1, err
	}
	envs, err := target.NewEnv(nil, &target.RawConfig{Envs: envsConfig}, target.ParseOSEnvs(os.Environ()), c.W)
	if err != nil {
		return 1, err
	}
	var latestURL string
	var ok bool
	e, err := envs.Config()
	if err != nil {
		return 1, err
	}
	if latestURL, ok = e["LATEST_URL"]; !ok {
		return 1, fmt.Errorf("LATEST_URL env key not set")
	}

	err = loadLatestBinary(latestURL, os.TempDir(), binPath)
	if err != nil {
		return 1, err
	}

	// print out current version installed
	status, err := execute.Command(binPath+" version", c.W, c.WErr, nil)
	if status != 0 || err != nil {
		return status, err
	}

	return 0, nil
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"upgrade": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Upgrade the ron binary."
}
