package markdown

import (
	"github.com/jschaf/b2/pkg/markdown/parser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"io"
	"io/ioutil"
)

type Markdown struct {
	gm goldmark.Markdown
}

func New() *Markdown {
	gm := goldmark.New(goldmark.WithExtensions(parser.TOMLFrontmatter))
	return &Markdown{gm: gm}
}

func (m *Markdown) Parse(r io.Reader) (ast.Node, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return m.gm.Parser().Parse(text.NewReader(bs)), nil
}

func (m *Markdown) Render(w io.Writer, source []byte, n ast.Node) error {
	return m.gm.Renderer().Render(w, source, n)
}
