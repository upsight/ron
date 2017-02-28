package target

import (
	"path/filepath"
	"strings"
)

// File is a mapping of the config file to its parsed targets.
type File struct {
	rawConfig *RawConfig
	// Filepath is the path to the input file.
	Filepath string
	// Targets are the files targets.
	Targets map[string]*Target
}

// Basename will return the Filepath name of file without the extension.
func (f *File) Basename() string {
	basename := filepath.Base(f.Filepath)
	return strings.TrimSuffix(basename, filepath.Ext(basename))
}
