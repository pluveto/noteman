package mdreformatter

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestMathFormatting(t *testing.T) {
	for _, tc := range []struct {
		source, want string
		blocks       int
	}{
		{"$$x < y & z$$ After *math*.\n", "x &lt; y &amp; z\n$$\n</code></pre>\n\nAfter _math_.", 1},
		{"$$\nx=1\n$$\n$$\ny=2\n$$\n", "y=2\n$$", 2},
		{"- Item\n\n\t$$\n\tn=pq\n\t$$\n", "n=pq", 1},
		{`Before $x + \$5 < y$ after.`, `<code>$x + \$5 &lt; y$</code> after.`, 0},
		{"```text\n$$x$$\n```\n\n`$y$`", "$$x$$", 0},
	} {
		var lf bytes.Buffer
		if err := Format([]byte(tc.source), &lf, true); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(lf.String(), tc.want) {
			t.Errorf("want %q in %q", tc.want, lf.String())
		}
		if got := strings.Count(lf.String(), `<pre class="mathjax-preview">`); got != tc.blocks {
			t.Errorf("got %d blocks, want %d: %s", got, tc.blocks, lf.String())
		}
		var crlf bytes.Buffer
		if err := Format([]byte(strings.ReplaceAll(tc.source, "\n", "\r\n")), &crlf, true); err != nil {
			t.Fatal(err)
		}
		if lf.String() != crlf.String() {
			t.Errorf("line endings changed output: LF %q, CRLF %q", lf.String(), crlf.String())
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestFormatReturnsWriterError(t *testing.T) {
	if err := Format([]byte("$$x$$"), failingWriter{}, true); err == nil {
		t.Fatal("expected writer error")
	}
}
