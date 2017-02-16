package ron

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// File is a composite of os.File that has the path
// and contents of either a local or remote file.
type File struct {
	os.File
	Path     string
	Contents string
}

var (
	rexpURLString = `^((ftp|http|https):\/\/)+(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|((([0-9a-zA-Z\.]*-?[0-9a-zA-Z\.]*))|((www\.)?))?(([a-z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-z\x{00a1}-\x{ffff}]{2,}))?))(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`
	rexpURL       *regexp.Regexp

	// readAll ...
	readAll = func(inp io.Reader) (string, error) {
		data, err := ioutil.ReadAll(inp)
		return string(data), err
	}

	// loadFile ...
	loadFile = func(path string, inp io.Reader) (string, error) {
		if path == "" {
			return "", fmt.Errorf("path not set %s", path)
		}
		if isURL(path) {
			resp, err := http.Get(path)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			data, err := readAll(resp.Body)
			if resp.StatusCode != 200 {
				return "", fmt.Errorf("%d %s", resp.StatusCode, data)
			}
			return data, err
		}
		if inp == nil {
			data, err := ioutil.ReadFile(path)
			return string(data), err
		}
		return readAll(inp)
	}
)

func init() {
	rexpURL = regexp.MustCompile(rexpURLString)
}

// isURL checks if a path is a url.
// adapted from here https://godoc.org/github.com/asaskevich/govalidator#IsURL
func isURL(path string) bool {
	if path == "" || len(path) >= 2083 || len(path) <= 3 || strings.HasPrefix(path, ".") {
		return false
	}
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	return rexpURL.MatchString(path)
}

// NewFile ...
func NewFile(path string) (*File, error) {
	contents, err := loadFile(path, nil)
	if err != nil {
		return nil, err
	}
	f := &File{
		Contents: contents,
		Path:     path,
	}
	return f, err
}

// String returns the loaded files contents.
func (f *File) String() string {
	return f.Contents
}
