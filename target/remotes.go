package target

import (
	"io"

	"github.com/upsight/ron/execute"

	yaml "gopkg.in/yaml.v2"
)

// Remotes are the mapping of env to list of remote hosts.
type Remotes map[string][]*execute.SSHConfig

// List will return a list of all possible remotes defined
func (r *Remotes) List(w io.Writer) error {
	out, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
