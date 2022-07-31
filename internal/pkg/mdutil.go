package pkg

import (
	"path/filepath"

	"github.com/yuin/goldmark/ast"
)

func MdFindFirstHeading(root ast.Node) *ast.Node {
	if root.Kind() == ast.KindHeading && root.(*ast.Heading).Level == 1 {
		return &root
	}
	child := root.FirstChild()
	nchild := root.ChildCount()
	for nchild > 0 {
		ret := MdFindFirstHeading(child)
		if ret != nil {
			return ret
		}
		child = child.NextSibling()
		nchild--
	}
	return nil
}

func IsMarkdownExt(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".md" || ext == ".markdown" || ext == ".mdown" || ext == ".mdx"
}
