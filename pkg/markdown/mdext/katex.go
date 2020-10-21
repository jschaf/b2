package mdext

import (
	"fmt"
	"github.com/graemephi/goldmark-qjs-katex"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.uber.org/zap"
)

// katexTransformer adds the katex feature to the context if the document looks
// like it has TeX math.
type katexTransformer struct{}

func newKatexTransformer() *katexTransformer {
	return &katexTransformer{}
}

func (kt *katexTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == qjskatex.KindTex {
			mdctx.GetLogger(pc).Debug("Found katex node", zap.String("path", mdctx.GetFilePath(pc)))
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
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(newKatexTransformer(), 1200),
		),
	)
	ext := qjskatex.Extension{}
	ext.Extend(m)
}
