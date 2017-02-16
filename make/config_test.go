package make

import (
	"fmt"
	"path"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

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
	_, _, err := LoadConfigFile(path.Join(wrkdir, "testdata", "target_test.yaml"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMakeLoadConfigFilePathErr(t *testing.T) {
	_, _, err := LoadConfigFile(path.Join(wrkdir, "nothere.yaml"))
	if err == nil {
		t.Fatal("expected path err")
	}
}

func TestMakeLoadConfigFileYamlErr(t *testing.T) {
	_, _, err := LoadConfigFile(path.Join(wrkdir, "../target_test.go"))
	if err == nil {
		t.Fatal("expected path err")
	}
}

func TestMakeLoadConfigEmpty(t *testing.T) {
	_, _, err := LoadConfigFile(path.Join(wrkdir, "empty.yaml"))
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
c:
`)
	var c *EnvTargetConfig
	err := yaml.Unmarshal([]byte(config), &c)
	if err == nil {
		t.Error("yaml should be invalid")
	}
	got := extractConfigError("file.yaml", config, err)
	if got.Error() != want.Error() {
		t.Errorf("got:\n\"%s\"\nwant:\n\"%s\"", got, want)
	}
}
