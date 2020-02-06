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
	gm goldmark.Markdown
}

func New() *Markdown {
	gm := goldmark.New(goldmark.WithExtensions(mdext.NewTOMLFrontmatter()))
	return &Markdown{gm: gm}
}

func (m *Markdown) Parse(r io.Reader) (*PostAST, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ctx := parser.NewContext()

	node := m.gm.Parser().Parse(text.NewReader(bs), parser.WithContext(ctx))
	return &PostAST{
		Node: node,
		Meta: mdext.GetTOMLMeta(ctx),
	}, nil
}

func (m *Markdown) Render(w io.Writer, source []byte, p *PostAST) error {
	return m.gm.Renderer().Render(w, source, p.Node)
}
