package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncPreflightAndNoWriteBack(t *testing.T) {
	for _, invalid := range []string{"", "mathjax: invalid\n", "lang: []\n"} {
		t.Run(invalid, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			target := filepath.Join(root, "target")
			if err := os.MkdirAll(source, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(target, "zh"), 0755); err != nil {
				t.Fatal(err)
			}
			first := "---\ntitle: First\nslug: first\nlang: zh\ndate: '2024-01-01'\nmathjax: true\n---\n$$x=1$$\n"
			second := "---\ntitle: Second\nslug: second\ndate: '2024-01-01'\n" + invalid + "---\nText.\n"
			paths := []string{filepath.Join(source, "first.md"), filepath.Join(source, "second.md")}
			for i, data := range []string{first, second} {
				if err := os.WriteFile(paths[i], []byte(data), 0644); err != nil {
					t.Fatal(err)
				}
			}
			existing := filepath.Join(target, "zh", "first.md")
			if err := os.WriteFile(existing, []byte("existing published content"), 0644); err != nil {
				t.Fatal(err)
			}
			conf := &AppConf{Target: AppConfTarget{Mapping: map[string]string{filepath.ToSlash(source) + "/": filepath.ToSlash(target) + "/{{lang_prefix}}"}}}
			p := &SyncProcessor{appConf: conf, cmd: &SyncCmd{NoWriteBack: true}, tasks: paths}
			err := p.Execute()
			if (err != nil) != (invalid != "") {
				t.Fatalf("unexpected result: %v", err)
			}
			got, readErr := os.ReadFile(existing)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if invalid != "" {
				if string(got) != "existing published content" {
					t.Fatal("failed preflight overwrote published content")
				}
			} else if !strings.Contains(string(got), "x=1\n$$") {
				t.Errorf("missing formula in output: %s", got)
			}
			for i, want := range []string{first, second} {
				got, err := os.ReadFile(paths[i])
				if err != nil || string(got) != want {
					t.Fatalf("source changed: %q %v", paths[i], err)
				}
			}
		})
	}
}
