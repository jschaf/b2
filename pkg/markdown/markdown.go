package markdown

import (
	"fmt"
	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"io"
	"io/ioutil"

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
	Assets assets.Map
	// The full path to the markdown file that this AST represents.
	Path     string
	Features *mdctx.Features
}

// Global configuration options for parsing and rendering markdown.
type Options struct {
	CiteStyle    cite.Style
	CiteAttacher mdext.CitationReferencesAttacher
	// TOCStyle determines how to show the TOC. Defaults to not showing a TOC.
	TOCStyle           mdext.TOCStyle
	Extenders          []goldmark.Extender
	HeadingAnchorStyle mdext.HeadingAnchorStyle
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

// WithTOCStyle overrides the default TOC style. The default is to show no TOC.
func WithTOCStyle(s mdext.TOCStyle) Option {
	return func(m *Markdown) {
		m.opts.TOCStyle = s
	}
}

// WithCiteAttacher overrides the default attacher for references. The default
// attaches references before the end of the first article tag.
func WithCiteAttacher(c mdext.CitationReferencesAttacher) Option {
	return func(m *Markdown) {
		m.opts.CiteAttacher = c
	}
}

// WithHeadingAnchorStyle overrides the default content shown on hovering over
// a heading. The default is to show nothing on hover.
func WithHeadingAnchorStyle(s mdext.HeadingAnchorStyle) Option {
	return func(m *Markdown) {
		m.opts.HeadingAnchorStyle = s
	}
}

func WithExtender(e goldmark.Extender) Option {
	parser.WithAutoHeadingID()
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
		mdext.NewColonLineExt(),
		mdext.NewFootnoteExt(),
		mdext.NewHeaderExt(),
		mdext.NewHeadingExt(opts.HeadingAnchorStyle),
		mdext.NewHeadingIDExt(),
		mdext.NewImageExt(),
		mdext.NewKatexExt(),
		mdext.NewLinkExt(),
		mdext.NewParagraphExt(),
		mdext.NewSmallCapsExt(),
		mdext.NewTOCExt(opts.TOCStyle),
		mdext.NewTOMLExt(),
		mdext.NewTimeExt(),
		mdext.NewTypographyExt(),
		mdext.NewFigureExt(), // TODO: must come last, why?
	}
}

// New creates a new markdown parser and renderer allowing additional options
// beyond the defaults.
func New(l *zap.Logger, opts ...Option) *Markdown {
	m := &Markdown{
		logger: l,
		opts: Options{
			CiteStyle:          cite.IEEE,
			CiteAttacher:       mdext.NewCitationArticleAttacher(),
			TOCStyle:           mdext.TOCStyleNone,
			HeadingAnchorStyle: mdext.HeadingAnchorStyleNone,
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
	mdctx.SetFilePath(ctx, path)
	mdctx.SetRenderer(ctx, m.gm.Renderer())
	mdctx.SetLogger(ctx, m.logger)

	node := m.gm.Parser().Parse(text.NewReader(bs), parser.WithContext(ctx))
	if parseErrs := mdctx.PopErrors(ctx); len(parseErrs) == 1 {
		return nil, fmt.Errorf("parse errors in context: %w", parseErrs[0])
	} else if len(parseErrs) > 1 {
		return nil, fmt.Errorf("parse errors in context: %v", parseErrs)
	}
	meta := mdext.GetTOMLMeta(ctx)
	meta.Title = mdctx.GetTitle(ctx)
	mdAssets := mdctx.GetAssets(ctx)
	mdFeats := mdctx.GetFeatures(ctx)
	return &PostAST{
		Node:     node,
		Meta:     meta,
		Assets:   mdAssets,
		Path:     path,
		Source:   bs,
		Features: mdFeats,
	}, nil
}

func (m *Markdown) Render(w io.Writer, source []byte, p *PostAST) error {
	return m.gm.Renderer().Render(w, source, p.Node)
}
