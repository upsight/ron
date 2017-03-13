package target

import (
	"io"

	yaml "gopkg.in/yaml.v2"
)

// RemoteHost is an individual host configuration for creating
// and ssh connection.
type RemoteHost struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
	User string `json:"user" yaml:"user"`
}

// Remotes are the mapping of env to list of remote hosts.
type Remotes map[string][]*RemoteHost

// List will return a list of all possible remotes defined
func (r *Remotes) List(w io.Writer) error {
	out, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
