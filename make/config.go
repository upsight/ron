package make

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	fi "github.com/upsight/ron/file"
	template "github.com/upsight/ron/template"
)

// extractConfigError parses the error for line number and then
// generates the text surrounding it.
func extractConfigError(path, input string, inErr error) error {
	err := inErr
	re := regexp.MustCompile(`line ([0-9]+):.*$`)
	v := re.FindStringSubmatch(inErr.Error())
	if v != nil && len(v) > 1 {
		if lineNum, e := strconv.Atoi(v[1]); e == nil {
			text := []string{}
			inLines := strings.Split(input, "\n")
			between := [2]int{lineNum - 5, lineNum + 5}
			for i, line := range inLines {
				switch {
				case i+1 == lineNum:
					text = append(text, line+" <<<<<<<<<<")
				case i+1 > between[0] && i+1 < between[1]:
					text = append(text, line)
				}
			}
			err = fmt.Errorf("%s %s\n%s\n", path, inErr.Error(), strings.Join(text, "\n"))
		}
	}
	return err
}

// LoadConfigFile will open a given file path and return it's raw
// envs and targets.
var LoadConfigFile = func(path string) (string, string, error) {
	f, err := fi.NewFile(path)
	if err != nil {
		return "", "", err
	}
	content, err := template.RenderGo(path, f.String())
	if err != nil {
		return "", "", err
	}

	var c *EnvTargetConfig
	err = yaml.Unmarshal([]byte(content), &c)
	if err != nil {
		return "", "", extractConfigError(path, content, err)
	}
	if c == nil {
		return "", "", fmt.Errorf("empty file requires envs and target keys")
	}
	envs, err := yaml.Marshal(c.Envs)
	if err != nil {
		return "", "", err
	}
	targets, err := yaml.Marshal(c.Targets)
	if err != nil {
		return "", "", err
	}
	return string(envs), string(targets), nil
}

// LoadDefault loads the binary yaml file envs and targets.
var LoadDefault = func() error {
	defaultYaml, err := Asset("make/default.yaml")
	if err != nil {
		return err
	}
	content, err := template.RenderGo("builtin:make/default.yaml", string(defaultYaml))
	if err != nil {
		return err
	}

	var c *EnvTargetConfig
	err = yaml.Unmarshal([]byte(content), &c)
	if err != nil {
		return err
	}

	// load envs
	d, err := yaml.Marshal(c.Envs)
	if err != nil {
		return err
	}
	DefaultEnvConfig = string(d)

	// load targets
	d, err = yaml.Marshal(c.Targets)
	if err != nil {
		return err
	}
	DefaultTargetConfig = string(d)

	return nil
}
