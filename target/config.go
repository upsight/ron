package target

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	fi "github.com/upsight/ron/file"
	template "github.com/upsight/ron/template"
)

var (
	// DefaultTargets is what is in project root target/default.yaml if not specified.
	DefaultTargets string
	// DefaultEnvConfig is what is in project root target/default.yaml if not specified.
	DefaultEnvConfig string
)

// ConfigFile is used to unmarshal configuration files.
type ConfigFile struct {
	Envs    []map[string]string `json:"envs" yaml:"envs"`
	Targets map[string]struct {
		Before      []string `json:"before" yaml:"before"`
		After       []string `json:"after" yaml:"after"`
		Cmd         string   `json:"cmd" yaml:"cmd"`
		Description string   `json:"description" yaml:"description"`
	} `json:"targets" yaml:"targets"`
}

// EnvsString is used for debugging the loaded envs.
func (c *ConfigFile) EnvsString() string {
	envs, _ := yaml.Marshal(c.Envs)
	return string(envs)
}

// TargetsString is used for debugging the loaded targets.
func (c *ConfigFile) TargetsString() string {
	targets, _ := yaml.Marshal(c.Targets)
	return string(targets)
}

// RawConfig contains the raw strings from a loaded config file.
type RawConfig struct {
	Filepath string
	Envs     string
	Targets  string
}

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
			err = fmt.Errorf("%s %s\n%s ", path, inErr.Error(), strings.Join(text, "\n"))
		}
	}
	return err
}

func findConfigFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		fp := filepath.Join(dir, "ron.yaml")
		if _, err := os.Stat(fp); err == nil {
			return fp, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", nil
		}
		dir = parentDir
	}
}

// LoadConfigFiles loads the default and override config files and returns
// them as a slice. If defaultYamlPath is an empty string, the defaults
// compiled into ron will be used instead. If overrideYamlPath is blank,
// it will find the nearest parent folder containing a ron.yaml file and use
// that file instead. In that case, the path to that file will be returned
// so that the caller can change the working directory to that folder before
// running further commands.
func LoadConfigFiles(defaultYamlPath, overrideYamlPath string) ([]*RawConfig, string, error) {
	configs := []*RawConfig{}

	var err error
	defaultConfig := &RawConfig{
		Filepath: "builtin:target/default.yaml",
		Envs:     DefaultEnvConfig,
		Targets:  DefaultTargets,
	}
	if defaultYamlPath != "" {
		defaultConfig, err = LoadConfigFile(defaultYamlPath)
		if err != nil {
			return nil, "", err
		}
	}
	defaultConfig.Envs = strings.TrimSpace(defaultConfig.Envs)
	defaultConfig.Targets = strings.TrimSpace(defaultConfig.Targets)
	configs = append(configs, defaultConfig)

	foundConfigDir := ""
	if overrideYamlPath == "" {
		overrideYamlPath, err = findConfigFile()
		if err != nil {
			return nil, "", err
		}
		foundConfigDir = filepath.Dir(overrideYamlPath)
	}

	if overrideYamlPath != "" {
		overrideConfig, err := LoadConfigFile(overrideYamlPath)
		if err != nil {
			return nil, "", err
		}
		overrideConfig.Filepath = overrideYamlPath
		overrideConfig.Envs = strings.TrimSpace(overrideConfig.Envs)
		overrideConfig.Targets = strings.TrimSpace(overrideConfig.Targets)
		configs = append(configs, overrideConfig)
	}

	return configs, foundConfigDir, nil
}

// LoadConfigFile will open a given file path and return it's raw
// envs and targets.
var LoadConfigFile = func(path string) (*RawConfig, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	f, err := fi.NewFile(path)
	if err != nil {
		return nil, err
	}
	content, err := template.RenderGo(path, f.String())
	if err != nil {
		return nil, err
	}

	var c *ConfigFile
	err = yaml.Unmarshal([]byte(content), &c)
	if err != nil {
		return nil, extractConfigError(path, content, err)
	}
	if c == nil {
		return nil, fmt.Errorf("empty file requires envs and target keys")
	}
	envs, err := yaml.Marshal(c.Envs)
	if err != nil {
		return nil, err
	}
	targets, err := yaml.Marshal(c.Targets)
	if err != nil {
		return nil, err
	}
	return &RawConfig{Filepath: path, Envs: string(envs), Targets: string(targets)}, nil
}

// LoadDefault loads the binary yaml file envs and targets.
var LoadDefault = func() error {
	defaultYaml, err := Asset("target/default.yaml")
	if err != nil {
		return err
	}
	content, err := template.RenderGo("builtin:target/default.yaml", string(defaultYaml))
	if err != nil {
		return err
	}

	var c *ConfigFile
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
	DefaultTargets = string(d)

	return nil
}
