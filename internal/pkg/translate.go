package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// translate.googleapis.com/?client=gtx is now consistently 429'd as
// "automated queries". Use Chrome's dict endpoint first, then MyMemory.
var httpClient = &http.Client{Timeout: 10 * time.Second}

const translateUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func Translate(sourceText, sourceLang, targetLang string) (string, error) {
	sourceText = strings.Replace(sourceText, "c++", "cpp", -1)
	sourceText = strings.Replace(sourceText, "C++", "cpp", -1)
	if strings.TrimSpace(sourceText) == "" {
		return "", errors.New("empty source text")
	}
	if sourceLang == "" {
		sourceLang = "auto"
	}
	if targetLang == "" {
		targetLang = "en"
	}

	text, errGoogle := translateGoogle(sourceText, sourceLang, targetLang)
	if errGoogle == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	text, errMyMemory := translateMyMemory(sourceText, sourceLang, targetLang)
	if errMyMemory == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	return "", fmt.Errorf("all translation backends failed: google: %v; mymemory: %v", errGoogle, errMyMemory)
}

func translateGoogle(sourceText, sourceLang, targetLang string) (string, error) {
	u := "https://clients5.google.com/translate_a/t?client=dict-chrome-ex" +
		"&sl=" + url.QueryEscape(sourceLang) +
		"&tl=" + url.QueryEscape(targetLang) +
		"&q=" + url.QueryEscape(sourceText)

	body, err := getTranslate(u)
	if err != nil {
		return "", err
	}
	return parseGoogleTranslateResponse(body)
}

func translateMyMemory(sourceText, sourceLang, targetLang string) (string, error) {
	if sourceLang == "auto" {
		sourceLang = "zh"
	}
	u := "https://api.mymemory.translated.net/get?q=" + url.QueryEscape(sourceText) +
		"&langpair=" + url.QueryEscape(sourceLang+"|"+targetLang)

	body, err := getTranslate(u)
	if err != nil {
		return "", err
	}
	return parseMyMemoryResponse(body)
}

func getTranslate(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", translateUserAgent)
	req.Header.Set("Accept", "application/json")

	r, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, errors.New("error reading response body")
	}
	if r.StatusCode != http.StatusOK {
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 180 {
			snippet = snippet[:180] + "..."
		}
		return nil, fmt.Errorf("http %d: %s", r.StatusCode, snippet)
	}
	return body, nil
}

// parseGoogleTranslateResponse accepts both:
//
//	["translated text"]                          // clients5 dict-chrome-ex
//	[[["translated", "original", ...], ...], ...] // legacy gtx
func parseGoogleTranslateResponse(body []byte) (string, error) {
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return "", errors.New("error unmarshaling google translate data")
	}
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		return "", errors.New("no translated data in response")
	}

	if s, ok := arr[0].(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return "", errors.New("empty translation")
		}
		return s, nil
	}

	inner, ok := arr[0].([]interface{})
	if !ok {
		return "", errors.New("unexpected google translate response shape")
	}
	var parts []string
	for _, slice := range inner {
		innerArr, ok := slice.([]interface{})
		if !ok || len(innerArr) == 0 {
			continue
		}
		if s, ok := innerArr[0].(string); ok {
			parts = append(parts, s)
		}
	}
	text := strings.Join(parts, "")
	if strings.TrimSpace(text) == "" {
		return "", errors.New("no translated data in response")
	}
	return text, nil
}

type myMemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	ResponseStatus int `json:"responseStatus"`
}

func parseMyMemoryResponse(body []byte) (string, error) {
	var resp myMemoryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", errors.New("error unmarshaling mymemory data")
	}
	if resp.ResponseStatus != 0 && resp.ResponseStatus != http.StatusOK {
		return "", fmt.Errorf("mymemory status %d", resp.ResponseStatus)
	}
	text := strings.TrimSpace(resp.ResponseData.TranslatedText)
	if text == "" {
		return "", errors.New("no translated data in response")
	}
	if strings.HasPrefix(strings.ToUpper(text), "MYMEMORY WARNING") {
		return "", errors.New(text)
	}
	return text, nil
}
