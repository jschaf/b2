package markdown

import (
	"io"
	"io/ioutil"

	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type PostAST struct {
	Node ast.Node
	Meta mdext.PostMeta
}

type Markdown struct {
	src []byte
	gm  goldmark.Markdown
}

func New() *Markdown {
	gm := goldmark.New(
		goldmark.WithExtensions(
			mdext.NewTOMLExt(),
			mdext.NewArticleExt(),
			mdext.NewTimeExt(),
		),
		goldmark.WithExtensions())
	return &Markdown{gm: gm}
}

func (m *Markdown) Parse(r io.Reader) (*PostAST, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	m.src = bs
	ctx := parser.NewContext()

	node := m.gm.Parser().Parse(text.NewReader(bs), parser.WithContext(ctx))
	meta := mdext.GetTOMLMeta(ctx)
	meta.Title = mdext.GetTitle(ctx)
	return &PostAST{
		Node: node,
		Meta: meta,
	}, nil
}

func (m *Markdown) Render(w io.Writer, source []byte, p *PostAST) error {
	return m.gm.Renderer().Render(w, source, p.Node)
}

func (m *Markdown) extractTitle(node ast.Node) string {
	if node.FirstChild().Kind() == ast.KindHeading {
		return string(node.FirstChild().Text(m.src))
	}
	node.NextSibling()
	return ""
}
