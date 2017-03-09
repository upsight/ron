// Package target implements gnu make like target processing from yaml files
// and OS environment variables.
package target

import "fmt"

// MSS is an alias for map[string]string
type MSS map[string]string

// Make runs targets...like make kinda
type Make struct {
	Configs *Configs
}

// NewMake creates a Make type with config embedded.
func NewMake(configs *Configs) (*Make, error) {
	m := &Make{
		Configs: configs,
	}
	return m, nil
}

// Run executes the given target name.
func (m *Make) Run(names ...string) error {
	for _, name := range names {
		target, ok := m.Configs.Target(name)
		if !ok {
			return fmt.Errorf("%s target not found", name)
		}
		status, out, err := target.Run()
		if status != 0 || err != nil {
			return fmt.Errorf("%d %s %s", status, out, err)
		}
	}
	return nil
}
