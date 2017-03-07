package target

import (
	"os"
	"os/user"
	"testing"
)

func Test_splitTarget(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"", args{"a:b"}, "a", "b"},
		{"", args{":b"}, "", "b"},
		{"", args{"b"}, "", "b"},
		{"", args{"b:"}, "b", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitTarget(tt.args.target)
			equals(t, tt.want, got)
			equals(t, tt.want1, got1)
		})
	}
}

func Test_keyIn(t *testing.T) {
	type inTest struct {
		name   string
		inKey  string
		inKeys []string
		out    bool
	}

	var inTests = []inTest{
		{"", "a", []string{"a", "b", "c"}, true},
		{"", "x", []string{"a", "b", "c"}, false},
		{"", "a", []string{"b", "z", "y", "a"}, true},
		{"", "a", []string{"b", "z", "a", "y"}, true},
		{"", "x", []string{}, false},
		{"", "", []string{}, false},
	}
	for _, tt := range inTests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyIn(tt.inKey, tt.inKeys)
			equals(t, tt.out, got)
		})
	}
}

func Test_homeDir(t *testing.T) {
	// test without $HOME set which tries to cd && pwd
	u, err := user.Current()
	ok(t, err)
	equals(t, u.HomeDir, homeDir())
	// test with $HOME set
	err = os.Setenv("HOME", "/home/test")
	ok(t, err)
	equals(t, "/home/test", homeDir())
}
