package mdext

import (
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
	ID     string
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
}

// Possible states for parsing citations.
type citeParseState = int

const (
	citeSearch   citeParseState = iota // looking for [
	citeStart                          // after parsing [
	citeFoundKey                       // after parsing @
)

// citeSpan is the start and end span that contain a citation.
// We don't track offsets because we rely on the fact that the brackets are
// always in text inline with length 1.
type citeSpan struct {
	start, end *ast.Text
}

// Transform extracts all citations into Citation nodes.
func (ca citationASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	spans, err := ca.findSpans(node, reader)
	if err != nil {
		panic(err)
	}

	for _, span := range spans {
		ca.reparentCitationsSpan(span)
	}
}

func (ca citationASTTransformer) reparentCitationsSpan(span citeSpan) {
	if span.start.Segment.Len() != 1 || span.end.Segment.Len() != 1 {
		// This assumption holds because the link parser doesn't merge the text
		// segments back together after parsing [ and ].
		panic("expected start and stop to be single element segments " +
			"containing '[' and ']'")
	}
	p := span.start.Parent()
	c := NewCitation()
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
}

func (ca citationASTTransformer) findSpans(node *ast.Document, reader text.Reader) ([]citeSpan, error) {
	state := citeSearch
	spans := make([]citeSpan, 0)
	var start *ast.Text

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
			start = nil
			return ast.WalkContinue, nil

		case ast.TypeInline:
			if n.Kind() != ast.KindText {
				return ast.WalkSkipChildren, nil
			}
		}

		txt := n.(*ast.Text)

		bytes := txt.Text(reader.Source())
		for _, b := range bytes {
			switch b {
			case '[':
				state = citeStart
				start = txt

			case '@':
				if state == citeStart {
					state = citeFoundKey
				}

			case ']':
				if state == citeFoundKey {
					span := citeSpan{
						start: start,
						end:   txt,
					}
					spans = append(spans, span)
				}
			}
		}
		return ast.WalkContinue, nil
	})

	return spans, err
}

type citationRenderer struct{}

func (c citationRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCitation, c.render)
}

func (c citationRenderer) render(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = writer.WriteString("<cite>")
	} else {
		_, _ = writer.WriteString("</cite>")
	}
	return ast.WalkContinue, nil
}

type CitationExt struct{}

func NewCitationExt() *CitationExt {
	return &CitationExt{}
}

func (sc *CitationExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(citationASTTransformer{}, 99)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(citationRenderer{}, 999)))
}
