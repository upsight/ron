package target

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/upsight/ron/color"
	fi "github.com/upsight/ron/file"
	template "github.com/upsight/ron/template"
)

const (
	// ConfigDirName is the name of the folder where ron will look for yaml config
	// files.
	ConfigDirName = ".ron"
	// ConfigFileName is the main ron config file that overrides other files.
	ConfigFileName = "ron.yaml"
)

var (
	// DefaultYamlBytes is used for the default built in binary targets
	// exported into default.go. If using as a library you can override this
	// by setting target.DefaultYamlBytes globally during startup.
	//   defaultYaml, err := config.Asset("myconfig/default.yaml")
	//   if err != nil {
	//       return err
	//   }
	//   target.DefaultYamlBytes = defaultYaml
	DefaultYamlBytes = []byte{}
)

func init() {
	defaultYaml, err := Asset("target/default.yaml")
	if err != nil {
		fmt.Println(color.Red(err.Error()))
	} else {
		DefaultYamlBytes = defaultYaml
	}
}

// ConfigFile is used to unmarshal configuration files.
type ConfigFile struct {
	Envs    []map[string]string `json:"envs" yaml:"envs"`
	Remotes *Remotes            `json:"remotes" yaml:"remotes"`
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

// RemotesString is used for debugging the loaded remotes.
func (c *ConfigFile) RemotesString() string {
	remotes, _ := yaml.Marshal(c.Remotes)
	return string(remotes)
}

// RawConfig contains the raw strings from a loaded config file.
type RawConfig struct {
	Filepath string
	Envs     string
	Remotes  string
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
				case i+2 == lineNum:
					text = append(text, line+" <<<<<<<<<<")
				case i+2 > between[0] && i+2 < between[1]:
					text = append(text, line)
				}
			}
			err = fmt.Errorf("%s %s\n%s ", path, inErr.Error(), strings.Join(text, "\n"))
		}
	}
	return err
}

// findConfigFile will search for a ron.yaml file, starting with the current directory
// and then searching parent directories for a first occurrence.
func findConfigFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		fp := filepath.Join(dir, ConfigFileName)
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

// findConfigDirs will search in the current directory for a .ron folder
// with *.yaml files, and then search parent directories.
func findConfigDirs(curdir string, withHomeDirectory bool) (dirs []string) {
	if withHomeDirectory {
		defer func() {
			// append the users home directory before returning
			hd := filepath.Join(homeDir(), ConfigDirName)
			if _, err := os.Stat(hd); err == nil {
				dirs = append(dirs, hd)
			}
		}()
	}

	for {
		dirpath := filepath.Join(curdir, ConfigDirName)
		if _, err := os.Stat(dirpath); err == nil {
			dirs = append(dirs, dirpath)
			return
		}
		parentDir := filepath.Dir(curdir)
		if parentDir == curdir {
			return
		}
		curdir = parentDir
	}
}

// findConfigDirFiles will find any *.yaml files in a list of .ron directories.
func findConfigDirFiles(dirs []string) []string {
	files := []string{}
	for _, dir := range dirs {
		found, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
		if err == nil {
			files = append(files, found...)
		}
	}
	return files
}

// addRonDirConfigs will first find any .ron folders in the current
// directory, followed by appending the home directory .ron folder.
// Any errors here will abort adding conigs and just return.
func addRonDirConfigs(wd string, configs *[]*RawConfig, withHomeDirectory bool) {
	dirs := findConfigDirs(wd, withHomeDirectory)
	files := findConfigDirFiles(dirs)
	for _, file := range files {
		conf, err := LoadConfigFile(file)
		if err != nil {
			fmt.Println(color.Red(err.Error()))
			continue
		}
		*configs = append(*configs, conf)
	}
}

// addRonYamlFile will prepend the list of configs with
// any ron.yaml files that are found along with returning its location.
// oYamlPath is the path to the override yaml file.
func addRonYamlFile(oYamlPath string, configs *[]*RawConfig) (string, error) {
	var err error
	foundConfigDir := ""
	if oYamlPath == "" {
		oYamlPath, err = findConfigFile()
		if err != nil {
			return "", err
		}
		foundConfigDir = filepath.Dir(oYamlPath)
	}

	if oYamlPath != "" {
		oConfig, err := LoadConfigFile(oYamlPath)
		if err != nil {
			fmt.Println(color.Red(err.Error()))
			return "", err
		}
		oConfig.Filepath = oYamlPath
		oConfig.Envs = strings.TrimSpace(oConfig.Envs)
		oConfig.Remotes = strings.TrimSpace(oConfig.Remotes)
		oConfig.Targets = strings.TrimSpace(oConfig.Targets)
		// prepend the override config
		*configs = append([]*RawConfig{oConfig}, *configs...)
	}

	return foundConfigDir, err
}

// addDefaultYamlFile will add a default config which should always be
// last in priority. If no path option is given a built in default will
// be created. dYamlPath is the path to the default yaml file.
func addDefaultYamlFile(dYamlPath string, configs *[]*RawConfig) {
	envs, targets, err := BuiltinDefault()
	dConfig := &RawConfig{
		Filepath: "builtin:target/default.yaml",
		Envs:     envs,
		Targets:  targets,
	}
	if dYamlPath != "" {
		dConfig, err = LoadConfigFile(dYamlPath)
		if err != nil {
			fmt.Println(color.Red(err.Error()))
			return
		}
		dConfig.Filepath = dYamlPath
	}
	dConfig.Envs = strings.TrimSpace(dConfig.Envs)
	dConfig.Remotes = strings.TrimSpace(dConfig.Remotes)
	dConfig.Targets = strings.TrimSpace(dConfig.Targets)
	*configs = append(*configs, dConfig)
}

// LoadConfigFiles loads the default, override, and any directory config files
// and returns them as a slice. If defaultYamlPath is an empty string, the defaults
// compiled into ron will be used instead. If overrideYamlPath is blank,
// it will find the nearest parent folder containing a ron.yaml file and use
// that file instead. In that case, the path to that file will be returned
// so that the caller can change the working directory to that folder before
// running further commands.
func LoadConfigFiles(defaultYamlPath, overrideYamlPath string, withHomeDirectory bool) ([]*RawConfig, string, error) {
	configs := []*RawConfig{}

	foundConfigDir, err := addRonYamlFile(overrideYamlPath, &configs)

	wd, err := os.Getwd()
	if err == nil {
		addRonDirConfigs(wd, &configs, withHomeDirectory)
	}
	addDefaultYamlFile(defaultYamlPath, &configs)
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
	remotes, err := yaml.Marshal(c.Remotes)
	if err != nil {
		return nil, err
	}
	targets, err := yaml.Marshal(c.Targets)
	if err != nil {
		return nil, err
	}
	return &RawConfig{
		Envs:     string(envs),
		Filepath: path,
		Remotes:  string(remotes),
		Targets:  string(targets),
	}, nil
}

// BuiltinDefault loads the binary yaml file and returns envs, targets, and any errors.
func BuiltinDefault() (string, string, error) {
	content, err := template.RenderGo("builtin:target/default.yaml", string(DefaultYamlBytes))
	if err != nil {
		return "", "", err
	}

	var c *ConfigFile
	err = yaml.Unmarshal([]byte(content), &c)
	if err != nil {
		return "", "", err
	}

	// load envs
	d, err := yaml.Marshal(c.Envs)
	if err != nil {
		return "", "", err
	}
	envs := string(d)

	// load targets
	d, err = yaml.Marshal(c.Targets)
	if err != nil {
		return "", "", err
	}
	targets := string(d)

	return envs, targets, nil
}
