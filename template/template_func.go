package template

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/upsight/ron/color"

	"github.com/upsight/ron/execute"
)

var (
	// FuncMap provides the available methods for rendering yaml templates.
	FuncMap = map[string]interface{}{
		"bash":       bash,
		"makeSlice":  makeSlice,
		"get":        get,
		"split":      strings.Split,
		"underscore": underscore,
	}
)

// makeSlice is a utility function for templates to create
// a slice.
func makeSlice(args ...interface{}) []interface{} {
	return args
}

// get a value or use the default if it is nil
func get(item interface{}, defaultVal interface{}) interface{} {
	if item == nil {
		return defaultVal
	}
	return item
}

// underscore all string hyphens to underscore
func underscore(item string) string {
	return strings.Replace(item, "-", "_", -1)
}

// bash shell command to run
func bash(cmd interface{}, returnStatus bool, returnOutput bool, returnErr bool) interface{} {
	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}
	status, err := execute.Command(fmt.Sprintf("%s", cmd), &stdOut, &stdErr, nil)
	if err != nil {
		log.Println(color.Red(err.Error()))
	}
	switch {
	case returnStatus:
		return status
	case returnOutput:
		return strings.TrimSpace(stdOut.String())
	case returnErr:
		return strings.TrimSpace(stdErr.String())
	}
	return ""
}
