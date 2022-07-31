package pkg

import (
	"testing"
)

func TestRemoveComment(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				s: "",
			},
			want: "",
		},
		{name: "single char a", args: args{s: "a"}, want: "a"},
		{name: "single char / ", args: args{s: "/"}, want: "/"},
		{name: "single char \" ", args: args{s: "\""}, want: "\""},
		{"complex",
			args{s: `{a:"\"//\"",
        // comment
        b:""}
        // comment
        // comment
        `}, `{a:"\"//\"",
        
        b:""}
        
        
        `},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveComment(tt.args.s); got != tt.want {
				t.Errorf("RemoveComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleGlob(t *testing.T) {
	out := []string{}
	SimpleGlob("../../examples", &out)
	for _, v := range out {
		println(v)
	}
}
