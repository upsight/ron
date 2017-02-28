// Package target implements gnu make like target processing from yaml files
// and OS environment variables.
package target

import (
	"fmt"
	"log"

	"github.com/upsight/ron/color"
)

// MSS is an alias for map[string]string
type MSS map[string]string

// Make runs targets...like make kinda
type Make struct {
	Env     *Env
	Configs *Configs
}

func init() {
	err := LoadDefault()
	if err != nil {
		log.Println(color.Red(err.Error()))
		return
	}
}

// NewMake creates a Make type with config embedded.
func NewMake(env *Env, targets *Configs) (*Make, error) {
	m := &Make{
		Env:     env,
		Configs: targets,
	}
	return m, nil
}

// Run executes the given target name.
func (m *Make) Run(names ...string) error {
	for _, name := range names {
		target, ok := m.Configs.Target(name)
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
