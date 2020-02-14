package mdext

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type codeBlockRenderer struct{}

func (c *codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, c.render)
}

func (c *codeBlockRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		_, _ = w.WriteString("<pre><code")
		language := n.Language(source)
		if language != nil {
			_, _ = w.WriteString(" class=\"lang-")
			_, _ = w.Write(language)
			_, _ = w.WriteString("\"")
		}
		_ = w.WriteByte('>')
		l := n.Lines().Len()

		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			_, _ = w.Write(line.Value(source))
		}
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func NewCodeBlockRenderer() *codeBlockRenderer {
	return &codeBlockRenderer{}
}

// codeBlockExt is a Goldmark extension to register the AST transformer and
// renderer
type codeBlockExt struct{}

func NewCodeBlockExt() *codeBlockExt {
	return &codeBlockExt{}
}

func (c *codeBlockExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewCodeBlockRenderer(), 999),
		),
	)
}
