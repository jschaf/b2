package mdext

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindTOC represents a TOC node.
var KindTOC = ast.NewNodeKind("TOC")

// TOC contains directives to format a table of contents.
// TOC nodes are created from the ColonLine parser.
type TOC struct {
	ast.BaseBlock
	Headings []*ast.Heading
}

func NewTOC() *TOC {
	return &TOC{}
}

func (c *TOC) Kind() ast.NodeKind {
	return KindTOC
}

func (c *TOC) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, nil, nil)
}

// tocTransformer adds heading entries to the TOC node.
type tocTransformer struct {
	// How many heading levels (1-based) to include in the TOC. If depth is 0,
	// defaults to 3. For example, a depth of 2 includes H1 and H2 headings in the
	// TOC.
	depth int
}

func newTOCTransformer() tocTransformer {
	return tocTransformer{}
}

func (t tocTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	depth := t.depth
	if t.depth == 0 {
		depth = 3
	}

	headings := make([]*ast.Heading, 0, 3*depth) // assume 3 headings per level
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindHeading {
			return ast.WalkContinue, nil
		}
		if n.Type() == ast.TypeInline {
			return ast.WalkSkipChildren, nil
		}
		h := n.(*ast.Heading)
		if h.Level <= depth {
			headings = append(headings, h)
		}
		return ast.WalkSkipChildren, nil
	})

	l := ast.NewList('.')
	for _, heading := range headings {
		l.AppendChild(l, heading)
	}
}

// tocRenderer is the HTML renderer for a TOC node.
type tocRenderer struct{}

func newTOCRenderer() tocRenderer {
	return tocRenderer{}
}

func (tr tocRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindTOC, renderTOC)
}

func renderTOC(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<div class=toc>")
	} else {
		w.WriteString("</div>")
	}
	return ast.WalkContinue, nil
}

type TOCExt struct{}

func NewTOCExt() goldmark.Extender {
	return TOCExt{}
}

func (T TOCExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(newTOCTransformer(), 1000)))
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(newTOCRenderer(), 1000)))
}
