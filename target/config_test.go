package target

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"os"

	yaml "gopkg.in/yaml.v2"
)

func mustLoadConfigFile(t *testing.T, path string, isDefault bool) *RawConfig {
	c, err := LoadConfigFile(path)
	ok(t, err)
	return c
}

func Test_findConfigFile(t *testing.T) {
	wd, err := os.Getwd()
	ok(t, err)
	expected := filepath.Join(filepath.Dir(wd), ConfigFileName)
	found, err := findConfigFile()
	ok(t, err)
	equals(t, expected, found)
}

func Test_findConfigDirs(t *testing.T) {
	wd, err := os.Getwd()
	ok(t, err)
	want := filepath.Join(wd, "testdata", ConfigDirName)

	dirs := findConfigDirs(filepath.Join(wd, "testdata"), false)

	found := false
	for _, dir := range dirs {
		if dir == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("wanted %s in dirs, got %+v", want, dirs)
	}
}

func Test_findConfigDirFiles(t *testing.T) {
	wd, err := os.Getwd()
	ok(t, err)
	want := []string{
		filepath.Join(wd, "testdata/.ron/default.yaml"),
		filepath.Join(wd, "testdata/.ron/empty.yaml"),
		filepath.Join(wd, "testdata/.ron/ron.yaml"),
	}

	dirs := []string{filepath.Join(wd, "testdata/.ron")}
	files := findConfigDirFiles(dirs)
	sort.Strings(files)
	equals(t, want, files)
}

func TestLoadConfigFiles(t *testing.T) {
	d, err := os.Getwd()
	ok(t, err)
	d = filepath.Dir(d)
	tests := []struct {
		name             string
		defaultYamlPath  string
		overrideYamlPath string
		expectedConfigs  []*RawConfig
		expectedFound    string
	}{
		{
			name:          "00 no override option finds parent ron.yaml",
			expectedFound: d,
			expectedConfigs: []*RawConfig{
				mustLoadConfigFile(t, "../ron.yaml", false),
				mustLoadConfigFile(t, "../.ron/docker.yaml", true),
				mustLoadConfigFile(t, "../.ron/go.yaml", true),
				mustLoadConfigFile(t, "./default.yaml", true),
			},
		},
		{
			name:             "01 override option",
			overrideYamlPath: "testdata/target_test.yaml",
			expectedFound:    "",
			expectedConfigs: []*RawConfig{
				mustLoadConfigFile(t, "testdata/target_test.yaml", false),
				mustLoadConfigFile(t, "../.ron/docker.yaml", true),
				mustLoadConfigFile(t, "../.ron/go.yaml", true),
				mustLoadConfigFile(t, "default.yaml", true),
			},
		},
		{
			name:            "02 default option",
			defaultYamlPath: "testdata/target_test.yaml",
			expectedFound:   d,
			expectedConfigs: []*RawConfig{
				mustLoadConfigFile(t, "../ron.yaml", false),
				mustLoadConfigFile(t, "../.ron/docker.yaml", true),
				mustLoadConfigFile(t, "../.ron/go.yaml", true),
				mustLoadConfigFile(t, "testdata/target_test.yaml", false),
			},
		},
		{
			name:             "03 default and override options",
			overrideYamlPath: "testdata/target_test.yaml",
			defaultYamlPath:  "testdata/target_test.yaml",
			expectedFound:    "",
			expectedConfigs: []*RawConfig{
				mustLoadConfigFile(t, "testdata/target_test.yaml", false),
				mustLoadConfigFile(t, "../.ron/docker.yaml", true),
				mustLoadConfigFile(t, "../.ron/go.yaml", true),
				mustLoadConfigFile(t, "testdata/target_test.yaml", false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs, found, err := LoadConfigFiles(tt.defaultYamlPath, tt.overrideYamlPath, false)
			ok(t, err)
			for i, config := range configs {
				equals(t, strings.TrimSpace(tt.expectedConfigs[i].Envs), strings.TrimSpace(config.Envs))
				equals(t, strings.TrimSpace(tt.expectedConfigs[i].Targets), strings.TrimSpace(config.Targets))
			}
			equals(t, tt.expectedFound, found)
		})
	}
}

func TestLoadConfigFileWithRemotes(t *testing.T) {
	c, err := LoadConfigFile(path.Join(wrkdir, "testdata", "ron.yaml"))
	ok(t, err)
	want := `production:
- host: exampleprod.com
  port: 22
  user: test
staging:
- host: example1.com
  port: 22
  user: test
- host: example2.com
  port: 22
  user: test
`
	equals(t, want, c.Remotes)
}

func TestLoadConfigFile(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "testdata", "target_test.yaml"))
	ok(t, err)
}

func TestLoadConfigFilePathErr(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "nothere.yaml"))
	if err == nil {
		t.Fatal("expected path err")
	}
}

func TestLoadConfigFileYamlErr(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "../target_test.go"))
	if err == nil {
		t.Fatal("expected path err")
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "testdata", "empty.yaml"))
	if err == nil {
		t.Fatalf(`expected error "empty file requires envs and target keys" got %v`, err)
	}
}

func TestExtractConfigError(t *testing.T) {

	config := `
envs:
  - APP: ron
  - APP: ron
  - APP: ron
  - APP: ron
  - APP: ron

{{abcd : }}
targets:
a:
b:
c:
d:
e:
f:`

	want := fmt.Errorf(`file.yaml yaml: line 10: could not find expected ':'
  - APP: ron
  - APP: ron
  - APP: ron

{{abcd : }} <<<<<<<<<<
targets:
a:
b:
c: `)
	var c *ConfigFile
	err := yaml.Unmarshal([]byte(config), &c)
	if err == nil {
		t.Error("yaml should be invalid")
	}
	got := extractConfigError("file.yaml", config, err)

	fmt.Println("want-----------------------------------------")
	fmt.Println(want)
	fmt.Println("-----------------------------------------")
	fmt.Println("got-----------------------------------------")
	fmt.Println(got)
	fmt.Println("-----------------------------------------")
	equals(t, want, got)
}
