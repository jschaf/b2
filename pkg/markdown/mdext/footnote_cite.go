package mdext

import (
	"errors"
	"fmt"
	"github.com/jschaf/b2/pkg/bibtex"
	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"strconv"
)

var KindCitation = ast.NewNodeKind("citation")

// Citation is an inline node representing a citation.
// See https://pandoc.org/MANUAL.html#citations.
type Citation struct {
	ast.BaseInline
	Key bibtex.CiteKey
	// The absolute path to the entry that contained this citation, like
	// "/til/qux".
	AbsPath string
	// The order number to use for this citation. Starts at 1. Updated by
	// FootnoteOrder called by the footnote order transformer.
	Order int
	// The bibtex entry this citation points to.
	Bibtex bibtex.Entry
	// The prefix in a citation reference, i.e the "foo" in `[foo @qux]`.
	Prefix string
	// The suffix in a citation reference, i.e the "bar" in `[@qux bar]`.
	Suffix string
	// The unique HTML ID of this citation.
	ID string
}

func (c *Citation) FootnoteOrder(nextOrder int, seen map[string]int, _ parser.Context) (FnAction, string) {
	seenOrder, ok := seen[c.Key]
	if ok {
		c.Order = seenOrder
		return FnOrderKeep, c.Key
	} else {
		c.Order = nextOrder
		return FnOrderNext, c.Key
	}
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
	if sc.citeStyle != cite.IEEE {
		panic("unsupported cite style: " + sc.citeStyle)
	}
	extenders.AddASTTransform(m, &citationParseTransformer{
		citeOrders:    make(map[bibtex.CiteKey]citeOrder),
		nextCiteOrder: 0,
		attacher:      sc.attacher,
	},
		ord.CitationTransformer)
	extenders.AddASTTransform(m, &citationIEEEFormatTransformer{}, ord.CitationFormatTransformer)
	extenders.AddRenderer(m, &citationRendererIEEE{
		includeRefs: sc.attacher != nil,
	}, ord.CitationRenderer)
}
