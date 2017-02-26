package target

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"os"

	yaml "gopkg.in/yaml.v2"
)

func mustLoadConfigFile(t *testing.T, path string, isDefault bool) *Config {
	c, err := LoadConfigFile(path)
	ok(t, err)
	return c
}

func TestMakeFindConfigFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(filepath.Dir(wd), "ron.yaml")
	found, err := findConfigFile()
	ok(t, err)
	equals(t, expected, found)
}

func TestMakeLoadConfigFiles(t *testing.T) {
	d, err := os.Getwd()
	ok(t, err)
	d = filepath.Dir(d)
	tests := []struct {
		name             string
		defaultYamlPath  string
		overrideYamlPath string
		expectedConfigs  []*Config
		expectedFound    string
	}{
		{
			name:          "",
			expectedFound: d,
			expectedConfigs: []*Config{
				mustLoadConfigFile(t, "default.yaml", true),
				mustLoadConfigFile(t, "../ron.yaml", false),
			},
		},
		{
			name:             "",
			overrideYamlPath: "testdata/target_test.yaml",
			expectedFound:    "",
			expectedConfigs: []*Config{
				mustLoadConfigFile(t, "default.yaml", true),
				mustLoadConfigFile(t, "testdata/target_test.yaml", false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("testing: %+v", tt)
			configs, found, err := LoadConfigFiles(tt.defaultYamlPath, tt.overrideYamlPath)
			ok(t, err)
			matched := true
			if len(configs) != len(tt.expectedConfigs) {
				matched = false
			}
			if !matched {
				t.Fail()
				t.Log("expected:\n")
				for _, c := range tt.expectedConfigs {
					t.Logf("- %+v\n", *c)
				}
				t.Log("got:\n")
				for _, c := range configs {
					t.Logf("- %+v\n", *c)
				}
			}
			equals(t, tt.expectedFound, found)
		})
	}
}

func TestMakeLoadDefaultAssetMissing(t *testing.T) {
	defaultAssetFunc, _ := _bindata["make/default.yaml"]
	defer func() {
		_bindata["make/default.yaml"] = defaultAssetFunc
	}()

	delete(_bindata, "make/default.yaml")
	err := LoadDefault()
	if err == nil {
		t.Fatal("expected missing config error")
	}
}

func TestMakeLoadConfigFile(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "testdata", "target_test.yaml"))
	ok(t, err)
}

func TestMakeLoadConfigFilePathErr(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "nothere.yaml"))
	if err == nil {
		t.Fatal("expected path err")
	}
}

func TestMakeLoadConfigFileYamlErr(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "../target_test.go"))
	if err == nil {
		t.Fatal("expected path err")
	}
}

func TestMakeLoadConfigEmpty(t *testing.T) {
	_, err := LoadConfigFile(path.Join(wrkdir, "testdata", "empty.yaml"))
	if err == nil {
		t.Fatalf(`expected error "empty file requires envs and target keys" got %v`, err)
	}
}

func TestMakeExtractConfigError(t *testing.T) {
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
	want := fmt.Errorf(`file.yaml yaml: line 9: could not find expected ':'
  - APP: ron
  - APP: ron
  - APP: ron

{{abcd : }} <<<<<<<<<<
targets:
a:
b:
c: `)
	var c *EnvTargetConfigs
	err := yaml.Unmarshal([]byte(config), &c)
	if err == nil {
		t.Error("yaml should be invalid")
	}
	got := extractConfigError("file.yaml", config, err)
	equals(t, want, got)
}
