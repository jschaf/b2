package mdext

import (
	"fmt"

	"github.com/graemephi/goldmark-qjs-katex"
	"github.com/jschaf/jsc/pkg/markdown/extenders"
	"github.com/jschaf/jsc/pkg/markdown/mdctx"
	"github.com/jschaf/jsc/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// katexTransformer adds the katex feature to the context if the document looks
// like it has TeX math.
type katexTransformer struct{}

func newKatexFeatureTransformer() *katexTransformer {
	return &katexTransformer{}
}

func (kt *katexTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == qjskatex.KindTex {
			mdctx.AddFeature(pc, mdctx.FeatureKatex)
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		mdctx.PushError(pc, fmt.Errorf("find katex nodes: %w", err))
	}
}

// KatexExt is a Goldmark extension to render TeX math using Katex.
type KatexExt struct{}

func NewKatexExt() *KatexExt {
	return &KatexExt{}
}

func (ke *KatexExt) Extend(m goldmark.Markdown) {
	extenders.AddASTTransform(m, newKatexFeatureTransformer(), ord.KatexFeatureTransformer)
	extenders.Extend(m, &qjskatex.Extension{}, int(ord.KatexParser), int(ord.KatexRenderer))
}
