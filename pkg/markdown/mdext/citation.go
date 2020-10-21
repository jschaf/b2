package mdext

import (
	"errors"
	"fmt"
	"github.com/jschaf/b2/pkg/bibtex"
	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/markdown/asts"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"strconv"
)

var KindCitation = ast.NewNodeKind("citation")

// Citation is an inline node representing a citation.
// See https://pandoc.org/MANUAL.html#citations.
type Citation struct {
	ast.BaseInline
	Key bibtex.CiteKey
	// The order that this citation appeared in the document, relative to other
	// citations. Starts at 0. The order always increments for each citation even
	// if preceding citations had the same key.
	Order int
	// The bibtex entry this citation points to.
	Bibtex bibtex.Entry
	// The prefix in a citation reference, i.e the "foo" in `[foo @qux]`.
	Prefix string
	// The suffix in a citation reference, i.e the "bar" in `[@qux bar]`.
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

// CiteID returns the HTML CiteID that links to a citation.
func (c *Citation) CiteID(count int) string {
	if count == 0 {
		return "cite_" + c.Key
	}
	return "cite_" + c.Key + "_" + strconv.Itoa(count)
}

// ReferenceID returns the HTML ID that links to the full reference for a
// citation, displayed in the reference section, if any.
func (c *Citation) ReferenceID() string {
	return "cite_ref_" + c.Key
}

type citationRenderer struct {
	citeStyle   cite.Style
	includeRefs bool
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
			nextNum:    1,
			citeNums:   make(map[bibtex.CiteKey]int),
			citeCounts: make(map[bibtex.CiteKey]int),
		},
	}
}

func (cr *citationRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	r, ok := citationRenderers()[cr.citeStyle]
	if !ok {
		panic("unsupported cite style: '" + cr.citeStyle + "'")
	}
	reg.Register(KindCitation, r.renderCitation)
	if cr.includeRefs {
		reg.Register(KindCitationReferences, r.renderReferenceList)
	} else {
		reg.Register(KindCitationReferences, asts.NopRender)
	}
}

// CitationReferencesAttacher determines how references are attached to the
// source document.
type CitationReferencesAttacher interface {
	Attach(doc *ast.Document, refs *CitationReferences) error
}

// CitationArticleAttacher attaches citation references to the end of an
// article.
type CitationArticleAttacher struct {
}

func NewCitationArticleAttacher() CitationArticleAttacher {
	return CitationArticleAttacher{}
}

func (c CitationArticleAttacher) Attach(doc *ast.Document, refs *CitationReferences) error {
	var article *Article
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		if _, ok := n.(*Article); !ok {
			return ast.WalkContinue, nil
		}
		article = n.(*Article)
		return ast.WalkStop, nil
	})
	if err != nil {
		return fmt.Errorf("citation: article attacher: %w", err)
	}
	if article == nil {
		return errors.New("citation: no article node found to attach citation references")
	}

	article.AppendChild(article, refs)
	return nil
}

func NewCitationNopAttacher() CitationReferencesAttacher {
	return nil
}

type CitationExt struct {
	citeStyle cite.Style
	attacher  CitationReferencesAttacher
}

func NewCitationExt(citeStyle cite.Style, attacher CitationReferencesAttacher) *CitationExt {
	return &CitationExt{citeStyle: citeStyle, attacher: attacher}
}

func (sc *CitationExt) Extend(m goldmark.Markdown) {
	extenders.AddASTTransform(
		m,
		&citationASTTransformer{
			citeStyle:     sc.citeStyle,
			citeOrders:    make(map[bibtex.CiteKey]citeOrder),
			nextCiteOrder: 0,
			attacher:      sc.attacher,
		},
		ord.CitationTransformer,
	)
	extenders.AddRenderer(m, &citationRenderer{
		citeStyle:   sc.citeStyle,
		includeRefs: sc.attacher != nil,
	}, ord.CitationRenderer)
}
