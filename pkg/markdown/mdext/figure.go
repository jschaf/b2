package mdext

import (
	"fmt"
	"log"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindFigure = ast.NewNodeKind("Figure")

type Figure struct {
	ast.BaseBlock
	Destination []byte
	Title       []byte
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

// figureASTTransformer converts a paragraph with a single image into a figure.
type figureASTTransformer struct{}

func (f *figureASTTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() != ast.KindParagraph {
			return ast.WalkContinue, nil
		}
		para := n.(*ast.Paragraph)
		if para.ChildCount() != 1 {
			return ast.WalkSkipChildren, nil
		}
		if para.FirstChild().Kind() != ast.KindImage {
			return ast.WalkSkipChildren, nil
		}
		img := para.FirstChild().(*ast.Image)

		fig := NewFigure()
		fig.Destination = img.Destination
		fig.Title = img.Title

		parent := para.Parent()
		if parent == nil {
			return ast.WalkSkipChildren, nil
		}
		parent.ReplaceChild(parent, para, fig)
		return ast.WalkSkipChildren, nil
	})

	if err != nil {
		log.Printf("error in paragraph transformer: %s", err)
	}
}

type figureRenderer struct{}

func (f *figureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFigure, f.render)
}

func (f *figureRenderer) render(w util.BufWriter, _ []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Figure)
	_, _ = w.WriteString("<figure>")
	_, _ = w.WriteString("<picture>")
	_, _ = w.WriteString(fmt.Sprintf("<img src=%q title=%q>", n.Destination, n.Title))
	_, _ = w.WriteString("</picture>")
	_, _ = w.WriteString("</figure>")

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
			util.Prioritized(&figureRenderer{}, 999)))
}
