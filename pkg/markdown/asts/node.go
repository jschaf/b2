// package asts contains utilities for working with Goldmark ASTs.
package asts

import (
	"github.com/yuin/goldmark/ast"
)

// Reparent moves all children of the src node to children of the dest node.
func Reparent(dest, src ast.Node) {
	cur := src.FirstChild()
	for cur != nil {
		next := cur.NextSibling()
		dest.AppendChild(dest, cur)
		cur = next
	}
}

// HeadingWalker is a function that will be called when WalkHeadings find a
// header. If HeadingWalker returns error, Walk function immediately
// stop walking.
type HeadingWalker func(n *ast.Heading) (ast.WalkStatus, error)

// WalkHeadings walks all headings and only calls walker when entering a
// heading.
func WalkHeadings(node ast.Node, walker HeadingWalker) error {
	return ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Type() == ast.TypeInline {
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() != ast.KindHeading {
			return ast.WalkContinue, nil
		}
		h := n.(*ast.Heading)
		return walker(h)
	})
}
