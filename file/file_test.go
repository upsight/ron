package ron

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

var (
	wrkdir       string
	configServer *httptest.Server
	mockLoadFile = func(path string, inp io.Reader) (string, error) {
		return testContents, nil
	}
	mockLoadFileErr = func(path string, inp io.Reader) (string, error) {
		return "", fmt.Errorf("nope")
	}
	testContents = `
a:
  b:
    - c
    - d
e:
  f: 1
`
)

func init() {
	configServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/notfound") {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, string(testContents))
	}))
	wrkdir, _ = os.Getwd()
}

func TestRonNewFile(t *testing.T) {
	prevLoadFile := loadFile
	defer func() { loadFile = prevLoadFile }()
	loadFile = mockLoadFile

	_, err := NewFile("/path/here")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRonNewFileLoadErr(t *testing.T) {
	prevLoadFile := loadFile
	defer func() { loadFile = prevLoadFile }()
	loadFile = mockLoadFileErr

	_, err := NewFile("/path/here")
	if err == nil {
		t.Fatal("expected err on file load")
	}
}

func TestRonFileString(t *testing.T) {
	prevLoadFile := loadFile
	defer func() { loadFile = prevLoadFile }()
	loadFile = mockLoadFile

	f, _ := NewFile("/path/here")
	if f.String() != testContents {
		t.Fatalf("expected %s got %s", testContents, f.String())
	}
}

func TestRonLoadFile(t *testing.T) {
	_, err := loadFile(path.Join(wrkdir, "file_test.go"), nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRonLoadFileURL(t *testing.T) {
	got, err := loadFile(configServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != testContents+"\n" {
		t.Fatalf("expected %s got %s", testContents, got)
	}
}

func TestRonLoadFileURLNotFound(t *testing.T) {
	_, err := loadFile(configServer.URL+"/notfound", nil)
	if err == nil {
		t.Fatal("expected err for not found url")
	}
}

func TestRonLoadFileIoReader(t *testing.T) {
	got, err := loadFile(wrkdir, strings.NewReader(testContents))
	if err != nil {
		t.Fatal(err)
	}
	if got != testContents {
		t.Fatalf("expected %s got %s", testContents, got)
	}
}

func TestRonLoadFileEmptyPath(t *testing.T) {
	_, err := loadFile("", nil)
	if err == nil {
		t.Fatal("expected err for empty path")
	}
}

func TestRonReadAll(t *testing.T) {
	got, err := readAll(strings.NewReader(testContents))
	if err != nil {
		t.Fatal(err)
	}
	if got != testContents {
		t.Fatalf("expected %s got %s", testContents, got)
	}
}

func TestRonIsURL(t *testing.T) {
	type tableTest struct {
		in  string
		out bool
	}

	tableTests := []tableTest{
		// valid urls
		{"http://foo.com/blah_blah", true},
		{"http://foo.com/blah_blah_(wikipedia)", true},
		{"http://✪df.ws/123", true},
		{"http://userid:password@example.com:8080", true},
		{"http://userid:password@example.com:8080/", true},
		{"http://➡.ws/䨹", true},
		{"http://223.255.255.254", true},
		{"http://-.~_!$&'()*+,;=:%40:80%2f::::::@example.com", true},
		{"http://مثال.إختبار", true},
		{"ftp://foo.bar/baz", true},
		{"http://142.42.1.1:8080/", true},
		{"http://142.42.1.1:8080/", true},
		{"http://127.0.0.1:8000/", true},
		{"http://localhost:8000/", true},
		{"http://localhost/", true},
		{"http://localhost", true},
		{"http://abcd", true},
		{"http://127.0.0.1", true},
		{"http://127.0.0.1/", true},
		{"http://a.b-c.de", true},
		{"http://a-b.b.com/path/a.yaml", true},
		{"http://a-b.b-d.com/path/a-b.yaml", true},
		// invalid urls
		{"", false},
		{"/abcd/", false},
		{"./a/b/c", false},
		{"../a/b/c.y", false},
		{"/a/b/c", false},
		{`/a/b/c\ d/boo`, false},
		{"http://foo.bar?q=Spaces should be encoded", false},
		{"http://", false},
		{"http://.", false},
		{"http://../", false},
		{"http://-error-.invalid/", false},
		{"ftps://foo.bar/", false},
	}
	for i, test := range tableTests {
		got := isURL(test.in)
		if got != test.out {
			t.Errorf("%d path: \"%s\" want %+v got %+v", i, test.in, test.out, got)
		}
	}
}
