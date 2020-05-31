package mdext

import (
	"fmt"
	"io/ioutil"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/cite/bibtex"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// citationASTTransformer extracts consecutive nodes that make up a citation
// from the AST and reparents the nodes as children of a new Citation node at
// the same position in the AST.
type citationASTTransformer struct {
	citeStyle cite.Style
	// The cite order for bibtex keys.
	citeOrders map[bibtex.Key]citeOrder
	// The next number to use for the raw citation order. Starts at 0.
	nextCiteOrder int
}

type citeOrder struct {
	key   bibtex.Key
	order int
	bib   bibtex.Element
}

// Possible states for parsing citations.
type citeParseState = int

const (
	citeSearch   citeParseState = iota // looking for '['
	citeStart                          // after parsing '['
	citeFoundKey                       // after parsing "@foobar"
	citeParseKey                       // after parsing "@foo" and hitting the end
)

// citeSpan is the start and end span that contain a citation.
type citeSpan struct {
	key        bibtex.Key
	order      int
	start, end *ast.Text
	// Absolute offsets that delimit the start and end of a citation.
	startOffset, endOffset int
}

// Transform extracts all citations into Citation nodes.
func (ca *citationASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	spans, err := ca.findSpans(doc, reader)
	if err != nil {
		PushError(pc, err)
		return
	}

	bibs := GetTOMLMeta(pc).BibPaths
	bibElems, err := ca.readBibs(bibs)
	if err != nil {
		PushError(pc, err)
		return
	}

	rl := NewReferenceList()
	for _, span := range spans {
		c, err := ca.newCitationParent(span, reader.Source())
		if err != nil {
			PushError(pc, err)
			return
		}
		bib, ok := bibElems[c.Key]
		if !ok {
			PushError(pc, fmt.Errorf("citation: no bibtex found for key: %s", c.Key))
			return
		}
		c.Bibtex = bib
		rl.Citations = append(rl.Citations, c)
	}
}

// readBibs returns all bibtex elements from the file paths in bibs merged into
// a map by the key.
func (ca *citationASTTransformer) readBibs(bibs []string) (map[bibtex.Key]*bibtex.Element, error) {
	bibEntries := make(map[string]*bibtex.Element)
	for _, bib := range bibs {
		bibBytes, err := ioutil.ReadFile(bib)
		if err != nil {
			return nil, fmt.Errorf("citation: read bib file: %w", err)
		}
		bibElems, err := bibtex.Parse(bibBytes)
		if err != nil {
			return nil, fmt.Errorf("citation: parse bib file: %w", err)
		}
		for _, elem := range bibElems {
			for _, key := range elem.Keys {
				bibEntries[key] = elem
			}
		}
	}
	return bibEntries, nil
}

// newCitationParent creates a citation node and reparents all spans between the
// start span to the end span inclusive as children of the newly created
// citation node.
func (ca *citationASTTransformer) newCitationParent(span citeSpan, source []byte) (*Citation, error) {
	p := span.start.Parent()
	// Split start and end nodes if there is other text besides the citation.
	if span.startOffset > span.start.Segment.Start {
		ss := span.start
		newStart := ast.NewText()
		newStart.Segment = ss.Segment.WithStart(span.startOffset)
		ss.Segment = ss.Segment.WithStop(span.startOffset)
		p.InsertAfter(p, ss, newStart)
		span.start = newStart
	}
	if span.endOffset < span.end.Segment.Stop {
		se := span.end
		newEnd := ast.NewText()
		newEnd.Segment = se.Segment.WithStop(span.endOffset)
		se.Segment = se.Segment.WithStart(span.endOffset)
		p.InsertBefore(p, se, newEnd)
		span.end = newEnd
	}

	// Remove the brackets.
	ss := span.start
	se := span.end
	ss.Segment = ss.Segment.WithStart(ss.Segment.Start + 1)
	se.Segment = se.Segment.WithStop(se.Segment.Stop - 1)

	// Reparent all spans between start and end inclusive.
	c := NewCitation()
	c.Key = span.key
	c.Order = span.order
	p.InsertBefore(p, span.start, c)
	var node ast.Node = span.start
	end := span.end.NextSibling()
	for node != end {
		cur := node
		node = node.NextSibling()
		c.AppendChild(c, cur)
	}

	return c, nil
}

// findSpans finds all bracketed citation spans, like [@foo, pp. 2].
func (ca *citationASTTransformer) findSpans(node *ast.Document, reader text.Reader) ([]citeSpan, error) {
	state := citeSearch
	startOffset := -1
	var start *ast.Text
	id := ""
	resetSearch := func() {
		state = citeSearch
		start = nil
		startOffset = -1
		id = ""
	}
	spans := make([]citeSpan, 0)
	order := 0

	// TODO: Drive our own walk function. It's hard to parse citations with event
	// dispatch parsing since we need to keep parsing state in-between function
	// calls.
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		// Skip everything except entering ast.Text. The brackets don't mean
		// anything in any other inline node, so don't go into the children of
		// inline nodes.
		if !entering {
			return ast.WalkContinue, nil
		}
		nodeType := n.Type()
		switch nodeType {
		case ast.TypeDocument, ast.TypeBlock:
			resetSearch()
			return ast.WalkContinue, nil

		case ast.TypeInline:
			if n.Kind() != ast.KindText {
				if state == citeParseKey {
					// If we hit another non-text node after starting to parse a bibtex
					// key, we finished parsing the key.
					state = citeFoundKey
				}
				return ast.WalkSkipChildren, nil
			}
		}

		txt := n.(*ast.Text)

		bytes := txt.Text(reader.Source())
		for i := 0; i < len(bytes); /* increment i manually */ {
			b := bytes[i]
			switch state {
			case citeSearch:
				if b == '[' {
					state = citeStart
					start = txt
					startOffset = txt.Segment.Start + i
				}

			case citeStart:
				switch b {
				case '@':
					i++
					lo := i
					for ; i < len(bytes) && bibtex.IsValidKeyChar(bytes[i]); i++ {
					}
					hi := i
					if hi > lo {
						id = string(bytes[lo:hi])
						state = citeFoundKey
						if i >= len(bytes) {
							// If we hit the end, the key might be over multiple spans.
							state = citeParseKey
						}
					}
					continue // don't increment, we already did above

				case '[':
					resetSearch()
					state = citeStart
				case ']':
					resetSearch()
				}

			case citeParseKey:
				lo := i
				for ; i < len(bytes) && bibtex.IsValidKeyChar(bytes[i]); i++ {
				}
				hi := i
				idSuffix := string(bytes[lo:hi])
				id = id + idSuffix
				state = citeFoundKey
				if i >= len(bytes) {
					// If we hit the end, the key might be over multiple different spans.
					state = citeParseKey
				}
				continue // don't increment, we already did above

			case citeFoundKey:
				switch b {
				case ']':
					span := citeSpan{
						key:         id,
						start:       start,
						order:       order,
						end:         txt,
						startOffset: startOffset,
						endOffset:   txt.Segment.Start + i + 1,
					}
					order++
					spans = append(spans, span)
					resetSearch()
				}
			}
			// If we didn't short-circuit, increment i.
			i++
		}
		return ast.WalkContinue, nil
	})

	return spans, err
}
