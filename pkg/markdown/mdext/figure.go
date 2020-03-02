package mdext

import (
	"path"
	"strings"

	"github.com/jschaf/b2/pkg/markdown/asts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	KindFigure     = ast.NewNodeKind("Figure")
	KindFigCaption = ast.NewNodeKind("FigCaption")
)

type Figure struct {
	ast.BaseBlock
	Destination []byte
	Title       []byte
	AltText     []byte
}

func NewFigure() *Figure {
	return &Figure{}
}

func (f *Figure) Dump(source []byte, level int) {
	ast.DumpHelper(f, source, level, nil, nil)
}

func (f *Figure) Kind() ast.NodeKind {
	return KindFigure
}

// FigCaption represents the caption for a figure, a `<figcaption>` in HTML5.
type FigCaption struct {
	ast.BaseBlock
}

func NewFigCaption() *FigCaption {
	return &FigCaption{}
}

func (f *FigCaption) Kind() ast.NodeKind {
	return KindFigCaption
}

func (f *FigCaption) Dump(source []byte, level int) {
	ast.DumpHelper(f, source, level, nil, nil)
}

// figureASTTransformer converts a paragraph with a single image into a figure.
type figureASTTransformer struct{}

const FigureCaptionMarker = "CAPTION:"

func isSingleImgParagraph(n *ast.Paragraph) bool {
	return n.ChildCount() == 1 && n.FirstChild().Kind() == ast.KindImage
}
func isCaption(n ast.Node, r text.Reader) bool {
	if n == nil || n.Kind() != ast.KindParagraph || n.FirstChild() == nil ||
		n.FirstChild().Kind() != ast.KindText {
		return false
	}
	s := string(n.FirstChild().Text(r.Source()))
	prefix := strings.HasPrefix(s, FigureCaptionMarker)
	return prefix
}

func (f *figureASTTransformer) Transform(doc *ast.Document, r text.Reader, pc parser.Context) {
	// Extract all single para images.
	imgs := make([]*ast.Image, 0, 4)
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		switch n.Kind() {
		case ast.KindParagraph:
			if isSingleImgParagraph(n.(*ast.Paragraph)) {
				img := n.FirstChild().(*ast.Image)
				imgs = append(imgs, img)
			}
			return ast.WalkSkipChildren, nil
		default:
			return ast.WalkContinue, nil
		}
	})
	if err != nil {
		panic(err)
	}

	// Replace each image with a figure.
	figs := make([]*Figure, 0, 4)
	for _, img := range imgs {
		fig := NewFigure()
		urlPath := GetTOMLMeta(pc).Path
		dest := path.Join(urlPath, string(img.Destination))
		fig.Destination = []byte(dest)
		fig.Title = img.Title
		fig.AltText = img.Text(r.Source())

		para := img.Parent()
		root := para.Parent()
		root.ReplaceChild(root, para, fig)
		figs = append(figs, fig)
	}

	// Pull captions into the figure if they have the appropriate marker.
	for _, fig := range figs {
		capt := fig.NextSibling()
		if !isCaption(capt, r) {
			continue
		}
		txt := capt.FirstChild().(*ast.Text)
		txt.Segment.Start += len(FigureCaptionMarker)
		figCaption := NewFigCaption()
		asts.Reparent(figCaption, capt)
		parent := capt.Parent()
		parent.RemoveChild(parent, capt)
		fig.AppendChild(fig, figCaption)
	}
}

// figureRenderer renders a Figure type.
type figureRenderer struct {
	html.Config
}

func NewFigureRenderer() renderer.NodeRenderer {
	return &figureRenderer{
		Config: html.NewConfig(),
	}
}

func (f *figureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFigure, f.renderFigure)
	reg.Register(KindFigCaption, f.renderFigCaption)
}

func (f *figureRenderer) renderFigure(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*Figure)
	if entering {
		_, _ = w.WriteString("<figure>")
		_, _ = w.WriteString("<picture>")
		_, _ = w.WriteString("<img src=\"")
		escapedSrc := util.EscapeHTML(util.URLEscape(n.Destination, true))
		_, _ = w.Write(escapedSrc)
		_, _ = w.WriteString("\" alt=\"")
		_, _ = w.Write(n.AltText)
		_ = w.WriteByte('"')
		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			f.Writer.Write(w, n.Title)
			_ = w.WriteByte('"')
		}
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.ImageAttributeFilter)
		}
		_, _ = w.WriteString(">")
		_, _ = w.WriteString("</picture>")
	} else {
		_, _ = w.WriteString("</figure>")
	}
	return ast.WalkContinue, nil
}

func (f *figureRenderer) renderFigCaption(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<figcaption>")
	} else {
		_, _ = w.WriteString("</figcaption>")
	}
	return ast.WalkContinue, nil
}

type figureExt struct{}

func NewFigureExt() *figureExt {
	return &figureExt{}
}

func (f *figureExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&figureASTTransformer{}, 999)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewFigureRenderer(), 999)))
}
