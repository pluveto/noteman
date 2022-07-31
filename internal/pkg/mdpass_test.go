package pkg

import (
	"testing"
)

func TestExtractMarkdownMeta(t *testing.T) {
	type args struct {
		source []rune
	}
	tests := []struct {
		name    string
		args    args
		wantRet *MarkdownMetaBody
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				source: []rune{},
			},
			wantRet: &MarkdownMetaBody{},
		},
		{
			name: "empty_meta",
			args: args{
				source: []rune(`---
---`),
			},
			wantRet: &MarkdownMetaBody{},
		},
		{
			name: "unterminated_meta",
			args: args{
				source: []rune(`---
--`),
			},
			wantRet: &MarkdownMetaBody{RawBody: `---
--`},
		},
		{
			name: "pure_text",
			args: args{
				source: []rune("hello world"),
			},
			wantRet: &MarkdownMetaBody{
				RawBody: "hello world",
			},
		},
		{
			name: "meta",
			args: args{
				source: []rune(`       
				---
a: 1
b: 2
c: 你好😀
---
hello world`),
			},
			wantRet: &MarkdownMetaBody{
				RawMeta: `a: 1
b: 2
c: 你好😀
`,
				RawBody: "hello world",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := ExtractMarkdownMeta(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractMarkdownMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			l := MustJsonEncode(gotRet)
			r := MustJsonEncode(tt.wantRet)
			if l != r {
				t.Errorf("ExtractMarkdownMeta() = %v, want %v", l, r)
			}
		})
	}
}
