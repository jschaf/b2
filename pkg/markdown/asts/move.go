package asts

import "github.com/yuin/goldmark/ast"

// Reparent moves all children of the src node to children of the dest node.
func Reparent(dest, src ast.Node) {
	cur := src.FirstChild()
	for cur != nil {
		next := cur.NextSibling()
		dest.AppendChild(dest, cur)
		cur = next
	}

}
