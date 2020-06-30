package asts

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

func WriteSlugText(dest []byte, node ast.Node, src []byte) []byte {
	offs := appendSlugText(dest, 0, node, src)
	return dest[:offs]
}

// WriteSlugText writes the node text content recursively into w using URL safe
// characters, namely ASCII letters and digits.
func appendSlugText(dest []byte, offs int, node ast.Node, src []byte) int {
	if offs >= len(dest) {
		return offs
	}

	var bs []byte
	switch x := node.(type) {
	case *ast.String:
		bs = x.Value
	case *ast.Text:
		bs = x.Segment.Value(src)
	case *parser.Delimiter:
		bs = x.Segment.Value(src)
	default:
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			offs = appendSlugText(dest, offs, c, src)
		}
		return offs
	}

	inInvalidRun := false
	remaining := len(dest) - offs
	if len(bs) < remaining {
		remaining = len(bs)
	}
	for _, b := range bs[:remaining] {
		switch {
		case isValidSlugChar(b):
			inInvalidRun = false
			dest[offs] = b
			offs++
		case inInvalidRun:
			continue
		default:
			p := offs - 1
			if p >= 0 && dest[p] != '-' && dest[p] != '_' {
				dest[offs] = '-'
				offs++
			}
			inInvalidRun = true
		}
	}
	return offs
}

func isValidSlugChar(b byte) bool {
	return ('0' <= b && b <= '9') ||
		('a' <= b && b <= 'z') ||
		('A' <= b && b <= 'Z')
}
