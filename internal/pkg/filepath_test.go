package pkg

import "testing"

func TestBaseNoExt(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, ""},
		{".xx", args{".xxx"}, ""},
		{"a.md", args{"a.md"}, "a"},
		{"/a.md", args{"/a.md"}, "a"},
		{"/", args{"/"}, ""},
		{"/a", args{"/a"}, "a"},
		{"/a/b", args{"/a/b"}, "b"},
		{"/a/b.c", args{"/a/b.c"}, "b"},
		{"/a/.c", args{"/a/.c"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BaseNoExt(tt.args.fileName); got != tt.want {
				t.Errorf("BaseNoExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
