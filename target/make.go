// Package target implements gnu make like target processing from yaml files
// and OS environment variables.
package target

import (
	"fmt"
	"sync"

	"github.com/upsight/ron/color"
	"github.com/upsight/ron/execute"
)

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
		if len(m.Configs.RemoteHosts) > 0 {
			wg := &sync.WaitGroup{}
			for _, h := range m.Configs.RemoteHosts {
				wg.Add(1)
				go func(host *execute.SSHConfig) {
					defer wg.Done()
					status, out, err := target.RunRemote(host)
					if status != 0 || err != nil {
						msg := fmt.Sprintf("%s] %d %s %v\n", host.Host, status, out, err)
						m.Configs.StdErr.Write([]byte(color.Red(msg)))
						return
					}
				}(h)
			}
			wg.Wait()
		} else {
			status, out, err := target.Run()
			if status != 0 || err != nil {
				return fmt.Errorf("%d %s %v", status, out, err)
			}
		}
	}
	return nil
}
