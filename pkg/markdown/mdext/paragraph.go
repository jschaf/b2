package mdext

import (
	"github.com/jschaf/b2/pkg/markdown/attrs"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// paragraphRenderer is an HTML renderer for paragraphs that supports using a
// tag other than <p> if customTagAttr is set.
type paragraphRenderer struct {
}

func (p paragraphRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindParagraph, p.renderParagraph)
}

func (p paragraphRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Paragraph)
	tag := "p"
	if customTag := attrs.GetStringAttr(n, attrs.CustomTagAttr); customTag != "" {
		tag = customTag
	}
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<" + tag)
			html.RenderAttributes(w, n, html.ParagraphAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<" + tag + ">")
		}
	} else {
		_, _ = w.WriteString("</" + tag + ">\n")
	}
	return ast.WalkContinue, nil

}

type ParagraphExt struct {
}

func NewParagraphExt() ParagraphExt {
	return ParagraphExt{}
}

func (p ParagraphExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(paragraphRenderer{}, 10)))
}
