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
	Node   ast.Node
	Meta   mdext.PostMeta
	Source []byte
	// A map where the key is the destination path relative to the public dir.
	// The value is the absolute file path of an asset like an image.
	// For example, 1 entry might be ./img.png -> /home/joe/blog/img.png.
	Assets map[string]string
	// The full path to the markdown file that this AST represents.
	Path string
}

type Markdown struct {
	gm goldmark.Markdown
}

func defaultExtensions() []goldmark.Extender {
	return []goldmark.Extender{
		mdext.NewArticleExt(),
		mdext.NewCodeBlockExt(),
		mdext.NewHeaderExt(),
		mdext.NewImageExt(),
		mdext.NewFigureExt(),
		mdext.NewTimeExt(),
		mdext.NewTOMLExt(),
	}
}

func New(exts ...goldmark.Extender) *Markdown {
	gm := goldmark.New(
		goldmark.WithExtensions(exts...),
		goldmark.WithExtensions(defaultExtensions()...))
	return &Markdown{gm: gm}
}

func (m *Markdown) Parse(path string, r io.Reader) (*PostAST, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ctx := parser.NewContext()
	mdext.SetFilePath(ctx, path)

	node := m.gm.Parser().Parse(text.NewReader(bs), parser.WithContext(ctx))
	meta := mdext.GetTOMLMeta(ctx)
	meta.Title = mdext.GetTitle(ctx)
	assets := mdext.GetAssets(ctx)
	return &PostAST{
		Node:   node,
		Meta:   meta,
		Assets: assets,
		Path:   path,
		Source: bs,
	}, nil
}

func (m *Markdown) Render(w io.Writer, source []byte, p *PostAST) error {
	return m.gm.Renderer().Render(w, source, p.Node)
}
