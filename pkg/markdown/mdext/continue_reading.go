package mdext

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindContinueReading = ast.NewNodeKind("Continue Reading")

// ContinueReading is a block node representing the point at which to truncate
// post text for a list view of posts as on an index page.
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

const ContinueReadingText = "CONTINUE_READING"

// contReadingParser parses a block of the continue reading node.
// Any paragraph containing only CONTINUE_READING is converted to a continue
// reading node.
type contReadingParser struct{}

func (c contReadingParser) Trigger() []byte {
	return nil
}

func (c contReadingParser) Open(_ ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if bytes.HasPrefix(line, []byte(ContinueReadingText)) {
		meta := GetTOMLMeta(pc)
		return NewContinueReading("/" + meta.Slug), parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (c contReadingParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}

func (c contReadingParser) Close(ast.Node, text.Reader, parser.Context) {
}

func (c contReadingParser) CanInterruptParagraph() bool {
	return false
}

func (c contReadingParser) CanAcceptIndentedLine() bool {
	return false
}

// contReadingTransformer removes all nodes after the continue reading node.
type contReadingTransformer struct{}

func (c contReadingTransformer) Transform(doc *ast.Document, r text.Reader, _ parser.Context) {

	var contReading ast.Node

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != KindContinueReading {
			return ast.WalkContinue, nil
		}
		contReading = n
		return ast.WalkStop, nil
	})

	if contReading == nil {
		return
	}
	parent := contReading.Parent()
	if parent == nil {
		return
	}
	for contReading.NextSibling() != nil {
		parent.RemoveChild(parent, contReading.NextSibling())
	}
}

// nopContReadingRenderer doesn't render the continue reading node.
type nopContReadingRenderer struct {
}

func (n nopContReadingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindContinueReading, n.renderNopContReading)
}

func (n nopContReadingRenderer) Transform(*ast.Document, text.Reader, parser.Context) {
}

func (n nopContReadingRenderer) renderNopContReading(util.BufWriter, []byte, ast.Node, bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

// contReadingRenderer renders the continue reading node.
type contReadingRenderer struct {
	html.Config
}

func (c contReadingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindContinueReading, c.render)
}

func contReadingLink(link string) string {
	return `<a class="continue-reading" href="` + link + `">` +
		`<svg aria-hidden="true" focusable="false" data-prefix="fas" data-icon="book" class="svg-inline--fa fa-book fa-w-14" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512"><path d="M448 360V24c0-13.3-10.7-24-24-24H96C43 0 0 43 0 96v320c0 53 43 96 96 96h328c13.3.0 24-10.7 24-24v-16c0-7.5-3.5-14.3-8.9-18.7-4.2-15.4-4.2-59.3.0-74.7 5.4-4.3 8.9-11.1 8.9-18.6zM128 134c0-3.3 2.7-6 6-6h212c3.3.0 6 2.7 6 6v20c0 3.3-2.7 6-6 6H134c-3.3.0-6-2.7-6-6v-20zm0 64c0-3.3 2.7-6 6-6h212c3.3.0 6 2.7 6 6v20c0 3.3-2.7 6-6 6H134c-3.3.0-6-2.7-6-6v-20zm253.4 250H96c-17.7.0-32-14.3-32-32 0-17.6 14.4-32 32-32h285.4c-1.9 17.1-1.9 46.9.0 64z"></path></svg>` +
		`<div class="continue-reading-text">Continue reading</div></a>`
}

func (c contReadingRenderer) render(
	w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ContinueReading)
		link := contReadingLink(n.Link)
		_, _ = w.WriteString(link)
	}
	return ast.WalkContinue, nil
}

// ContinueReadingExt extends markdown with support to show the continue reading
// block and truncate all nodes after the continue reading block
type ContinueReadingExt struct{}

func NewContinueReadingExt() ContinueReadingExt {
	return ContinueReadingExt{}
}

func (c ContinueReadingExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(contReadingParser{}, 800)))

	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(contReadingTransformer{}, 1001)))

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(contReadingRenderer{}, 500)))
}

// NopContinueReadingExt extends markdown to ignore the continue reading block
type NopContinueReadingExt struct{}

func NewNopContinueReadingExt() NopContinueReadingExt {
	return NopContinueReadingExt{}
}

func (n NopContinueReadingExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(contReadingParser{}, 800)))

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(nopContReadingRenderer{}, 500)))
}
