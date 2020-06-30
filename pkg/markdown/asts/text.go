package asts

import (
	"bytes"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

const slugSep = '-'

// WriteSlugText writes the node text content recursively into dest using URL
// safe characters, namely lower case ASCII letters and digits. Applies the
// following transformations:
// - Converts all ASCII to lowercase.
// - Truncates the slug so it always ends with a complete word.
// - Removes trailing stop words like "the" and "a" but always keeps at least
//   two words.
func WriteSlugText(dest []byte, node ast.Node, src []byte) []byte {
	offs, isTruncated := appendSlugText(dest, 0, node, src)
	if isTruncated {
		for i := len(dest) - 1; i > 0; i-- {
			b := dest[i]
			if b == slugSep {
				offs = i
				break
			}
		}
	}
	return dropStopWords(dest[:offs])
}

func appendSlugText(dest []byte, offs int, node ast.Node, src []byte) (int, bool) {
	if offs >= len(dest) {
		return offs, false
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
		isTruncated := false
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			cOffs, cTrunc := appendSlugText(dest, offs, c, src)
			offs = cOffs
			isTruncated = isTruncated || cTrunc
		}
		return offs, isTruncated
	}

	inInvalidRun := false
	remaining := len(dest) - offs
	isTruncated := true
	if len(bs) <= remaining {
		isTruncated = false
		remaining = len(bs)
	}
	for _, b := range bs[:remaining] {
		switch {
		case isValidSlugChar(b):
			inInvalidRun = false
			dest[offs] = lower(b)
			offs++
		case inInvalidRun:
			continue
		default:
			p := offs - 1
			if p >= 0 && dest[p] != '-' && dest[p] != '_' {
				dest[offs] = slugSep
				offs++
			}
			inInvalidRun = true
		}
	}
	return offs, isTruncated
}

func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b | (1 << 5)
	}
	return b
}

func isValidSlugChar(b byte) bool {
	return ('0' <= b && b <= '9') ||
		('a' <= b && b <= 'z') ||
		('A' <= b && b <= 'Z') ||
		b == '.'
}

// Removes stop words at the end of a slug but always keeps at least 2 words.
// A word is contiguous ascii letters or digits.
func dropStopWords(dest []byte) []byte {
	stops := [...][]byte{
		[]byte("the"),
		[]byte("is"),
		[]byte("at"),
		[]byte("a"),
		[]byte("on"),
		[]byte("and"),
		[]byte("to"),
		[]byte("for"),
	}
	isStop := func(sub []byte) bool {
		for _, stop := range stops {
			if bytes.Equal(sub, stop) {
				return true
			}
		}
		return false
	}

	end := len(dest)
	wordCount := 0
	for i, b := range dest {
		if b == slugSep {
			wordCount++
			if wordCount == 2 {
				end = i
				break
			}
		}
	}

	offs := len(dest)
	for i := len(dest) - 1; i >= end; i-- {
		b := dest[i]
		if b == slugSep {
			sub := dest[i+1 : offs]
			if isStop(sub) {
				offs = i
				continue
			}
			break
		}
	}
	return dest[:offs]
}
