package pkg

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestParseGoogleTranslateResponse(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{
			name: "clients5 single string",
			body: `["What is moment generating function (MGF)?"]`,
			want: "What is moment generating function (MGF)?",
		},
		{
			name: "legacy gtx nested",
			body: `[[["Hello ","你好",null,null,1],["world","世界",null,null,1]]]`,
			want: "Hello world",
		},
		{
			name:    "html 429 body",
			body:    `<html><title>Sorry...</title></html>`,
			wantErr: true,
		},
		{
			name:    "empty array",
			body:    `[]`,
			wantErr: true,
		},
		{
			name:    "empty string translation",
			body:    `[""]`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGoogleTranslateResponse([]byte(tt.body))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseMyMemoryResponse(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			body: `{"responseData":{"translatedText":"Test title","match":0.99},"responseStatus":200}`,
			want: "Test title",
		},
		{
			name:    "quota warning",
			body:    `{"responseData":{"translatedText":"MYMEMORY WARNING: YOU USED ALL AVAILABLE FREE TRANSLATIONS FOR TODAY"},"responseStatus":200}`,
			wantErr: true,
		},
		{
			name:    "non-200",
			body:    `{"responseData":{"translatedText":""},"responseStatus":403}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMyMemoryResponse([]byte(tt.body))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTranslateHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`<html><title>Sorry...</title>automated queries</html>`))
	}))
	defer srv.Close()

	_, err := getTranslate(srv.URL)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("error should include status, got %v", err)
	}
}

func TestTranslateLive(t *testing.T) {
	if os.Getenv("NOTEMAN_LIVE") == "" {
		t.Skip("set NOTEMAN_LIVE=1 to run live translation")
	}
	got, err := Translate("什么是矩生成函数（MGF）？", "zh", "en")
	if err != nil {
		t.Fatalf("Translate live failed: %v", err)
	}
	if got == "" {
		t.Fatal("empty translation")
	}
	t.Logf("translated: %q", got)
}
