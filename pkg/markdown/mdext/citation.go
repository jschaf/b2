package mdext

import (
	"fmt"
	"io/ioutil"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/cite/bibtex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindCitation = ast.NewNodeKind("citation")

// Citation is an inline node representing a citation.
// See https://pandoc.org/MANUAL.html#citations.
type Citation struct {
	ast.BaseInline
	ID     bibtex.Key
	Bibtex *bibtex.Element
	Prefix string
	Suffix string
}

func NewCitation() *Citation {
	return &Citation{}
}

func (c Citation) Kind() ast.NodeKind {
	return KindCitation
}

func (c Citation) Dump(source []byte, level int) {
	ast.DumpHelper(&c, source, level, nil, nil)
}

type citationASTTransformer struct {
	citeStyle cite.Style
	// The cite order for bibtex keys.
	citeOrders map[bibtex.Key]int
	// The next number to use for IEEE type citation references, like [2].
	nextCiteOrder int
	bibElems      map[string]*bibtex.Element
}

// Possible states for parsing citations.
type citeParseState = int

const (
	citeSearch   citeParseState = iota // looking for [
	citeStart                          // after parsing [
	citeFoundKey                       // after parsing @foobar
	citeParseKey                       // after parsing @foo and hitting the end
)

// citeSpan is the start and end span that contain a citation.
// We don't track offsets because we rely on the fact that the brackets are
// always in text inline with length 1.
type citeSpan struct {
	id         string
	start, end *ast.Text
}

// Transform extracts all citations into Citation nodes.
func (ca *citationASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	spans, err := ca.findSpans(doc, reader)
	if err != nil {
		panic(err)
	}

	bibs := GetTOMLMeta(pc).BibPaths
	ca.bibElems, err = ca.readBibs(bibs)
	if err != nil {
		PushError(pc, err)
		return
	}

	for _, span := range spans {
		cSpan := ca.reparentCitationsSpan(span)
		if err := ca.styleCiteRefs(cSpan); err != nil {
			PushError(pc, err)
			return
		}
	}

	refNode := ca.buildReferences()
}

func (ca *citationASTTransformer) buildReferences() *ReferenceList {
	rl := NewReferenceList()
	refs := make([]*Reference, ca.nextCiteOrder-1)
	for key, order := range ca.citeOrders {
		refs[order] = key
	}
	for i := 0; i < ca.nextCiteOrder; i++ {

	}

}

// citeOrders returns the order that key appeared in the source document,
// starting at 1.
func (ca *citationASTTransformer) citeOrder(key bibtex.Key) int {
	if n, ok := ca.citeOrders[key]; ok {
		return n
	}
	n := ca.nextCiteOrder
	ca.nextCiteOrder++
	ca.citeOrders[key] = n
	return n
}

func (ca *citationASTTransformer) readBibs(bibs []string) (map[string]*bibtex.Element, error) {
	bibEntries := make(map[string]*bibtex.Element)

	for _, bib := range bibs {
		bibBytes, err := ioutil.ReadFile(bib)
		if err != nil {
			return nil, fmt.Errorf("citation AST transform read bib file: %w", err)
		}
		bibElems, err := bibtex.Parse(bibBytes)
		if err != nil {
			return nil, fmt.Errorf("citation AST transform parse bib file: %w", err)
		}
		for _, elem := range bibElems {
			for _, key := range elem.Keys {
				bibEntries[key] = elem
			}
		}
	}
	return bibEntries, nil
}

func (ca *citationASTTransformer) styleCiteRefs(c *Citation) error {
	bib, ok := ca.bibElems[c.ID]
	if !ok {
		return fmt.Errorf("style citation: no reference key found for ID: %s", c.ID)
	}
	c.Bibtex = bib

	n := fmt.Sprintf("[%d]", ca.citeOrder(c.ID))
	title := ast.NewString([]byte(n))
	c.InsertBefore(c, c.FirstChild(), title)
	return nil
}

func (ca *citationASTTransformer) reparentCitationsSpan(span citeSpan) *Citation {
	if span.start.Segment.Len() != 1 || span.end.Segment.Len() != 1 {
		// This assumption holds because the link parser doesn't merge the text
		// segments back together after parsing [ and ].
		panic("expected start and stop to be single element segments " +
			"containing '[' and ']'")
	}
	p := span.start.Parent()
	c := NewCitation()
	c.ID = span.id
	p.InsertBefore(p, span.start, c)
	var node = span.start.NextSibling()
	for node != span.end {
		cur := node
		node = node.NextSibling()
		c.AppendChild(c, cur)
	}
	// We don't care about the brackets.
	p.RemoveChild(p, span.start)
	p.RemoveChild(p, span.end)
	return c
}

// findSpans finds all bracketed citation spans, like [@foo, pp. 2].
func (ca *citationASTTransformer) findSpans(node *ast.Document, reader text.Reader) ([]citeSpan, error) {
	state := citeSearch
	var start *ast.Text
	id := ""
	resetSearch := func() {
		state = citeSearch
		start = nil
		id = ""
	}
	spans := make([]citeSpan, 0)

	// TODO: Drive our own walk function. Too hard to do this event dispatch based
	// parsing.
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
		for i := 0; i < len(bytes); i++ {
			b := bytes[i]
			switch state {
			case citeSearch:
				if b == '[' {
					state = citeStart
					start = txt
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

			case citeFoundKey:
				switch b {
				case ']':
					span := citeSpan{
						id:    id,
						start: start,
						end:   txt,
					}
					spans = append(spans, span)
					resetSearch()
				}
			}
			i++
		}
		return ast.WalkContinue, nil
	})

	return spans, err
}

type citationRenderer struct {
}

func (cr citationRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCitation, cr.render)
}

func (cr citationRenderer) render(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	c := n.(*Citation)
	if entering {
		_, _ = writer.WriteString(`<cite data-cite-key="`)
		_, _ = writer.WriteString(c.ID)
		_, _ = writer.WriteString(`">`)
	} else {
		_, _ = writer.WriteString("</cite>")
	}
	return ast.WalkContinue, nil
}

type CitationExt struct {
	citeStyle cite.Style
}

func NewCitationExt(citeStyle cite.Style) *CitationExt {
	return &CitationExt{citeStyle: citeStyle}
}

func (sc *CitationExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&citationASTTransformer{
				citeStyle:     sc.citeStyle,
				citeOrders:    make(map[bibtex.Key]int),
				nextCiteOrder: 1,
			}, 99)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(citationRenderer{}, 999)))
}
