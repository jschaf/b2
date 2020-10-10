package mdext

import (
	"github.com/graemephi/goldmark-qjs-katex"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// katexTransformer adds the katex feature to the context if the document looks
// like it has TeX math.
type katexTransformer struct{}

func newKatexTransformer() *katexTransformer {
	return &katexTransformer{}
}

func (kt *katexTransformer) Transform(_ *ast.Document, _ text.Reader, pc parser.Context) {
	// TODO: Only add the feature if we used katex.
	// https://github.com/graemephi/goldmark-qjs-katex/issues/7
	mdctx.AddFeature(pc, mdctx.FeatureKatex)
}

// KatexExt is a Goldmark extension to render TeX math using Katex.
type KatexExt struct{}

func NewKatexExt() *KatexExt {
	return &KatexExt{}
}

func (ke *KatexExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(newKatexTransformer(), 1200),
		),
	)
	ext := qjskatex.Extension{}
	ext.Extend(m)
}
