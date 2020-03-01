package mdext

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindContinueReading = ast.NewNodeKind("Continue Reading")

type ContinueReading struct {
	ast.BaseBlock
	Link string
}

func NewContinueReading(link string) *ContinueReading {
	return &ContinueReading{Link: link}
}

func (c *ContinueReading) Kind() ast.NodeKind {
	return KindContinueReading
}

func (c *ContinueReading) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, nil, nil)
}

const ContinueReadingText = "CONTINUE READING"

type contReadingTransformer struct{}

func NewContinueReadingTransformer() *contReadingTransformer {
	return &contReadingTransformer{}
}

func (c *contReadingTransformer) findSecondPara(doc *ast.Document) (*ast.Paragraph, error) {
	paraCount := 0
	var para *ast.Paragraph
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		switch n.Kind() {
		case ast.KindDocument:
			return ast.WalkContinue, nil
		case ast.KindParagraph:
			paraCount += 1
			if paraCount == 2 {
				para = n.(*ast.Paragraph)
				return ast.WalkStop, nil
			}
			return ast.WalkSkipChildren, nil
		default:
			return ast.WalkSkipChildren, nil
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find 1st paragraph: %w", err)
	}
	return para, nil
}

func findContReading(doc *ast.Document, reader text.Reader) (*ast.Paragraph, error) {
	var para *ast.Paragraph
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		switch n.Kind() {
		case ast.KindDocument:
			return ast.WalkContinue, nil
		case ast.KindParagraph:
			if isContinueReadingNode(n.(*ast.Paragraph), reader) {
				para = n.(*ast.Paragraph)
				return ast.WalkStop, nil
			}
			return ast.WalkSkipChildren, nil
		default:
			return ast.WalkSkipChildren, nil
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find 1st paragraph: %w", err)
	}
	return para, nil
}

func (c *contReadingTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	continueReading, err := findContReading(doc, reader)
	if err != nil {
		panic(err)
	}
	if continueReading == nil {
		continueReading, err = c.findSecondPara(doc)
		if err != nil {
			panic(err)
		}
	}
	if continueReading == nil {
		panic("failed to find continue reading node")
	}

	parent := continueReading.Parent()
	if parent == nil {
		panic("parent is nil")
	}
	if parent.Kind() != ast.KindDocument {
		panic("parent is not a document")
	}

	for continueReading.NextSibling() != nil {
		parent.RemoveChild(parent, continueReading.NextSibling())
	}
	parent.RemoveChild(parent, continueReading)
	meta := GetTOMLMeta(pc)
	parent.AppendChild(parent, NewContinueReading("/"+meta.Slug))
}

func isContinueReadingNode(n *ast.Paragraph, r text.Reader) bool {
	if n.ChildCount() != 1 || n.FirstChild().Kind() != ast.KindText {
		return false
	}
	txt := n.FirstChild().(*ast.Text)
	s := string(txt.Segment.Value(r.Source()))
	return s == ContinueReadingText
}

// nopContReadingTransformer removes the continue reading text if it exists.
type nopContReadingTransformer struct {
}

func (n *nopContReadingTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	r, err := findContReading(doc, reader)
	if err != nil {
		panic(err)
	}
	if r == nil {
		// Doesn't exist, okay to skip.
		return
	}
	p := r.Parent()
	p.RemoveChild(p, r)
}

type ContinueReadingRenderer struct {
	html.Config
}

func NewContinueReadingRenderer() *ContinueReadingRenderer {
	return &ContinueReadingRenderer{}
}

func (c *ContinueReadingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindContinueReading, c.render)
}

func contReadingLink(link string) string {
	return `<a class="continue-reading" href="` + link + `">` +
		`<svg aria-hidden="true" focusable="false" data-prefix="fas" data-icon="book" class="svg-inline--fa fa-book fa-w-14" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512"><path d="M448 360V24c0-13.3-10.7-24-24-24H96C43 0 0 43 0 96v320c0 53 43 96 96 96h328c13.3.0 24-10.7 24-24v-16c0-7.5-3.5-14.3-8.9-18.7-4.2-15.4-4.2-59.3.0-74.7 5.4-4.3 8.9-11.1 8.9-18.6zM128 134c0-3.3 2.7-6 6-6h212c3.3.0 6 2.7 6 6v20c0 3.3-2.7 6-6 6H134c-3.3.0-6-2.7-6-6v-20zm0 64c0-3.3 2.7-6 6-6h212c3.3.0 6 2.7 6 6v20c0 3.3-2.7 6-6 6H134c-3.3.0-6-2.7-6-6v-20zm253.4 250H96c-17.7.0-32-14.3-32-32 0-17.6 14.4-32 32-32h285.4c-1.9 17.1-1.9 46.9.0 64z"></path></svg>` +
		`<div class="continue-reading-text">Continue reading</div></a>`
}

func (c *ContinueReadingRenderer) render(
	w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ContinueReading)
		link := contReadingLink(n.Link)
		_, _ = w.WriteString(link)
	}
	return ast.WalkContinue, nil
}

type contReadingExt struct{}

func NewContinueReadingExt() *contReadingExt {
	return &contReadingExt{}
}

func (c *contReadingExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(NewContinueReadingTransformer(), 999)))

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewContinueReadingRenderer(), 500)))
}

type noContReadingExt struct{}

func NewNopContinueReadingExt() *noContReadingExt {
	return &noContReadingExt{}
}

func (n *noContReadingExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&nopContReadingTransformer{}, 999)))
}
