package markdown

import (
	"io"
	"io/ioutil"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.uber.org/zap"
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

// Global configuration options for parsing and rendering markdown.
type Options struct {
	CiteStyle    cite.Style
	CiteAttacher mdext.CitationReferencesAttacher
	Extenders    []goldmark.Extender
}

type Markdown struct {
	gm     goldmark.Markdown
	logger *zap.Logger
	opts   Options
}

// Option is a functional option that manipulates the Markdown struct.
type Option func(*Markdown)

// WithCiteStyle overrides the default citation style.
func WithCiteStyle(c cite.Style) Option {
	return func(m *Markdown) {
		m.opts.CiteStyle = c
	}
}

// WithCiteAttacher overrides the default attacher for references. The default
// is to attach the references to the end of the first article tag.
func WithCiteAttacher(c mdext.CitationReferencesAttacher) Option {
	return func(m *Markdown) {
		m.opts.CiteAttacher = c
	}
}

func WithExtender(e goldmark.Extender) Option {
	return func(m *Markdown) {
		m.opts.Extenders = append(m.opts.Extenders, e)
	}
}

func defaultExtensions(opts Options) []goldmark.Extender {
	return []goldmark.Extender{
		mdext.NewArticleExt(),
		mdext.NewCitationExt(opts.CiteStyle, opts.CiteAttacher),
		mdext.NewCodeBlockExt(),
		mdext.NewColonBlockExt(),
		mdext.NewHeaderExt(),
		mdext.NewImageExt(),
		mdext.NewLinkExt(),
		mdext.NewFigureExt(),
		mdext.NewSmallCapsExt(),
		mdext.NewTimeExt(),
		mdext.NewTOMLExt(),
		mdext.NewTypographyExt(),
	}
}

// New creates a new markdown parser and renderer with additional extenders
// beyond the default extenders.
func New(l *zap.Logger, opts ...Option) *Markdown {
	m := &Markdown{
		logger: l,
		opts: Options{
			CiteStyle:    cite.IEEE,
			CiteAttacher: mdext.NewCitationArticleAttacher(),
		},
	}
	for _, opt := range opts {
		opt(m)
	}

	m.gm = goldmark.New(
		goldmark.WithExtensions(m.opts.Extenders...),
		goldmark.WithExtensions(defaultExtensions(m.opts)...))
	return m
}

func (m *Markdown) Parse(path string, r io.Reader) (*PostAST, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ctx := parser.NewContext()
	mdext.SetFilePath(ctx, path)
	mdext.SetRenderer(ctx, m.gm.Renderer())
	mdext.SetLogger(ctx, m.logger)

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
