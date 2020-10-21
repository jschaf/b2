package extenders

import (
	"github.com/jschaf/b2/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func AddBlockParser(m goldmark.Markdown, p parser.BlockParser, pri ord.ParserPriority) {
	m.Parser().AddOptions(parser.WithBlockParsers(util.Prioritized(p, int(pri))))
}

func AddInlineParser(m goldmark.Markdown, p parser.InlineParser, pri ord.ParserPriority) {
	m.Parser().AddOptions(parser.WithInlineParsers(util.Prioritized(p, int(pri))))
}

func AddASTTransform(m goldmark.Markdown, t parser.ASTTransformer, pri ord.ASTTransformerPriority) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(t, int(pri))))
}

func AddRenderer(m goldmark.Markdown, t renderer.NodeRenderer, pri ord.RendererPriority) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(t, int(pri))))
}

// Keep varargs ints to allow putting in priorities as informational.
func Extend(m goldmark.Markdown, e goldmark.Extender, _ ...int) {
	e.Extend(m)
}
