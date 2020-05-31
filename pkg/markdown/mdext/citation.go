package mdext

import (
	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/cite/bibtex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var KindCitation = ast.NewNodeKind("citation")

// Citation is an inline node representing a citation.
// See https://pandoc.org/MANUAL.html#citations.
type Citation struct {
	ast.BaseInline
	Key bibtex.Key
	// The order that this citation appeared in the document, relative to other
	// citations. Starts at 0. The order always increments for each citation even
	// if the preceding citations had the same key.
	Order  int
	Bibtex *bibtex.Element
	Prefix string
	Suffix string
}

func NewCitation() *Citation {
	return &Citation{}
}

func (c *Citation) Kind() ast.NodeKind {
	return KindCitation
}

func (c *Citation) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, nil, nil)
}

// ID returns the HTML ID that links to a citation.
func (c *Citation) ID() string {
	return "cite_" + c.Key
}

// ReferenceID returns the HTML ID that links to the full reference for a
// citation, displayed in the reference section, if any.
func (c *Citation) ReferenceID() string {
	return "cite_ref_" + c.Key
}

type citationRenderer struct {
	citeStyle cite.Style
}

// citationStyleRenderer renders citations and references for a specific style.
// It's useful to have both renderers in a single struct in order to share
// information between the citation and reference list.
type citationStyleRenderer interface {
	renderCitation(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error)
	renderReferenceList(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error)
}

func citationRenderers() map[cite.Style]citationStyleRenderer {
	return map[cite.Style]citationStyleRenderer{
		cite.IEEE: &citationRendererIEEE{
			nextNum:  1,
			citeNums: make(map[bibtex.Key]int),
		},
	}
}

func (cr *citationRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	r, ok := citationRenderers()[cr.citeStyle]
	if !ok {
		panic("unsupported cite style: '" + cr.citeStyle + "'")
	}
	reg.Register(KindCitation, r.renderCitation)
	reg.Register(KindReferenceList, r.renderReferenceList)
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
				citeOrders:    make(map[bibtex.Key]citeOrder),
				nextCiteOrder: 0,
			}, 99)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&citationRenderer{
				citeStyle: sc.citeStyle,
			}, 999)))
}
