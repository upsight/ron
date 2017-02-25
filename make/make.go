// Package make implements gnu make like target processing from yaml files
// and OS environment variables.
package make

import (
	"fmt"
	"log"

	"github.com/upsight/ron/color"
)

var (
	// DefaultTarget is what is in root/make/default.yaml if not specified.
	DefaultTarget string
	// DefaultEnvConfig is what is in root/make/default.yaml if not specified.
	DefaultEnvConfig string
)

// MSS is an alias for map[string]string
type MSS map[string]string

// Make runs targets...like make kinda
type Make struct {
	Env          *Env
	TargetConfigs *TargetConfigs
}

// EnvTargetConfigs ...
type EnvTargetConfigs struct {
	Envs    []map[string]string `json:"envs" yaml:"envs"`
	Targets map[string]struct {
		Before      []string `json:"before" yaml:"before"`
		After       []string `json:"after" yaml:"after"`
		Cmd         string   `json:"cmd" yaml:"cmd"`
		Description string   `json:"description" yaml:"description"`
	} `json:"targets" yaml:"targets"`
}

func init() {
	err := LoadDefault()
	if err != nil {
		log.Println(color.Red(err.Error()))
		return
	}
}

// NewMake creates a Make type with config embedded.
func NewMake(env *Env, targets *TargetConfigs) (*Make, error) {
	m := &Make{
		Env:          env,
		TargetConfigs: targets,
	}
	return m, nil
}

// Run executes the given target name.
func (m *Make) Run(names ...string) error {
	for _, name := range names {
		target, ok := m.TargetConfigs.Target(name)
		if ok {
			status, out, err := target.Run()
			if status != 0 || err != nil {
				return fmt.Errorf("%d %s %s", status, out, err)
			}
		} else {
			return fmt.Errorf("%s target not found", name)
		}
	}
	return nil
}
