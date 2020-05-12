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

type continueReadingParser struct{}

func (c continueReadingParser) Trigger() []byte {
	return nil
}

func (c continueReadingParser) Open(_ ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if bytes.HasPrefix(line, []byte(ContinueReadingText)) {
		meta := GetTOMLMeta(pc)
		return NewContinueReading("/" + meta.Slug), parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (c continueReadingParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}

func (c continueReadingParser) Close(ast.Node, text.Reader, parser.Context) {
}

func (c continueReadingParser) CanInterruptParagraph() bool {
	return false
}

func (c continueReadingParser) CanAcceptIndentedLine() bool {
	return false
}

var defaultContinueReadingParser = &continueReadingParser{}

func NewContinueReadingParser() parser.BlockParser {
	return defaultContinueReadingParser
}

type contReadingTransformer struct{}

func NewContinueReadingTransformer() *contReadingTransformer {
	return &contReadingTransformer{}
}

func (c *contReadingTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	n := doc.FirstChild()
	if n == nil {
		return
	}
	for n != nil {
		if n.Kind() == KindContinueReading {
			break
		}
		n = n.NextSibling()
	}

	if n == nil {
		return
	}
	continueReading := n
	for continueReading.NextSibling() != nil {
		doc.RemoveChild(doc, continueReading.NextSibling())
	}
}

// nopContReadingRenderer removes the continue reading text if it exists.
type nopContReadingRenderer struct {
}

func (n *nopContReadingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindContinueReading, n.renderNopContReading)
}

func NewNopContinueReadingRenderer() *nopContReadingRenderer {
	return &nopContReadingRenderer{}
}

func (n *nopContReadingRenderer) Transform(*ast.Document, text.Reader, parser.Context) {
}

func (n *nopContReadingRenderer) renderNopContReading(util.BufWriter, []byte, ast.Node, bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
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
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewContinueReadingParser(), 800)))

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
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewContinueReadingParser(), 800)))

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewNopContinueReadingRenderer(), 500)))
}
