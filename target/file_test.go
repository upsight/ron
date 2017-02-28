package target

import "testing"

func TestFile_Basename(t *testing.T) {
	type fields struct {
		rawConfig *RawConfig
		Filepath  string
		Targets   map[string]*Target
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"", fields{nil, "/a/b/ron.yaml", nil}, "ron"},
		{"", fields{nil, "/a/b/default.yaml", nil}, "default"},
		{"", fields{nil, "default.yaml", nil}, "default"},
		{"", fields{nil, "", nil}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &File{
				rawConfig: tt.fields.rawConfig,
				Filepath:  tt.fields.Filepath,
				Targets:   tt.fields.Targets,
			}
			equals(t, tt.want, f.Basename())
		})
	}
}
