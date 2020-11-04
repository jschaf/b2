package extenders

import (
	"github.com/jschaf/b2/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// AddBlockParser adds the block parser to Goldmark markdown at the priority.
func AddBlockParser(m goldmark.Markdown, p parser.BlockParser, pri ord.ParserPriority) {
	m.Parser().AddOptions(parser.WithBlockParsers(util.Prioritized(p, int(pri))))
}

// AddInlineParser adds the inline parser to Goldmark markdown at the priority.
func AddInlineParser(m goldmark.Markdown, p parser.InlineParser, pri ord.ParserPriority) {
	m.Parser().AddOptions(parser.WithInlineParsers(util.Prioritized(p, int(pri))))
}

// AddASTTransform adds the AST transformer to Goldmark markdown at the
// priority.
func AddASTTransform(m goldmark.Markdown, t parser.ASTTransformer, pri ord.ASTTransformerPriority) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(t, int(pri))))
}

// AddParaTransform adds the paragraph AST transformer to Goldmark markdown at
// the priority.
func AddParaTransform(m goldmark.Markdown, t parser.ParagraphTransformer, pri ord.ParaTransformerPriority) {
	m.Parser().AddOptions(parser.WithParagraphTransformers(util.Prioritized(t, int(pri))))
}

// AddRenderer adds the renderer to Goldmark markdown at the priority.
func AddRenderer(m goldmark.Markdown, t renderer.NodeRenderer, pri ord.RendererPriority) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(t, int(pri))))
}

// Extend extends Goldmark markdown with the extender. Keep varargs ints to
// allow putting in priorities as informational.
func Extend(m goldmark.Markdown, e goldmark.Extender, _ ...int) {
	e.Extend(m)
}
