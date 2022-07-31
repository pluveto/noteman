package pkg

import "github.com/sirupsen/logrus"

type MarkdownMetaBody struct {
	RawMeta          string
	MetaChanged      bool
	Meta             map[string]interface{}
	RawBody          string
	RawBodyFormatted string
}

func (s *MarkdownMetaBody) DumpFormatted() string {
	return "---\n" + s.RawMeta + "---\n" + s.RawBodyFormatted
}
func (s *MarkdownMetaBody) Dump() string {
	return "---\n" + s.RawMeta + "---\n" + s.RawBody
}

func ExtractMarkdownMeta(source []rune) (ret *MarkdownMetaBody, err error) {
	ret = &MarkdownMetaBody{}
	pos := 0
	pos = eatWhitespace(source, pos)
	pos = eatKeyword(source, pos, "---")
	if pos < 0 {
		ret.RawBody = string(source)
		return
	}
	pos = eatWhitespace(source, pos)
	fmstart := pos
	pos = eatUntilKeyword(source, pos, []rune("---"))
	if pos < 0 {
		ret.RawBody = string(source)
		logrus.Warningln("bad meta detected at ", fmstart)
		return
	}
	ret.RawMeta = string(source[fmstart:pos])
	pos += len("---")
	pos = eatWhitespace(source, pos)
	ret.RawBody = string(source[pos:])
	return
}

func eatUntilKeyword(source []rune, pos int, pattern []rune) (nextpos int) {
	for pos < len(source) && !peekString(source, pos, pattern) {
		pos++
	}
	if pos >= len(source) {
		return -1
	}
	return pos
}

func eatKeyword(source []rune, pos int, keyword string) (nextpos int) {
	if !peekString(source, pos, []rune(keyword)) {
		return -1
	}
	return pos + len(keyword)
}
func peekString(source []rune, pos int, pattern []rune) bool {
	if pos+len(pattern) > len(source) {
		return false
	}
	for i, v := range pattern {
		if source[pos+i] != v {
			return false
		}
	}
	return true
}

func eatWhitespace(source []rune, pos int) (nextpos int) {
	for pos < len(source) {
		if source[pos] == ' ' || source[pos] == '\t' ||
			source[pos] == '\n' || source[pos] == '\r' {
			pos++
		} else {
			break
		}
	}
	return pos
}
