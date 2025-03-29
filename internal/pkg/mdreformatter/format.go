package mdreformatter

import (
	"io"

	// mathjax "github.com/litao91/goldmark-mathjax"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
)

// Format write reformatted markdown source.
//
// Use internal markdown parser with extensions GFM, DefinitionList,
// Footnote, LineBlocks, BlockAttributes and other.
func Format(source []byte, w io.Writer, math bool) error {
	extensions := []goldmark.Extender{
		extension.GFM,
		extension.DefinitionList,
		extension.Footnote,
	}
	if math {
		extensions = append(extensions, mathjax.MathJax)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(
			parser.WithAttribute(),
		),
	)
	doc := md.Parser().Parse(text.NewReader(source))
	return Render(w, source, doc)
}

// Markdown is a markdown format renderer.
var Markdown renderer.Renderer = new(markdownRenderer)

type markdownRenderer struct{}

// AddOptions adds given option to this renderer.
func (*markdownRenderer) AddOptions(opts ...renderer.Option) {}

// Write render node as Markdown.
func (*markdownRenderer) Render(w io.Writer, source []byte, node ast.Node) (err error) {
	return Render(w, source, node)
}
