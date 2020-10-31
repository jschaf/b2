package mdext

import (
	"github.com/jschaf/b2/pkg/markdown/attrs"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var KindCustomInline = ast.NewNodeKind("CustomInline")

type CustomInline struct {
	ast.BaseInline
	Tag string
}

func NewCustomInline(tag string) *CustomInline {
	return &CustomInline{
		Tag: tag,
	}
}

func (c *CustomInline) Kind() ast.NodeKind {
	return KindCustomInline
}

func (c *CustomInline) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, nil, nil)
}

type customInlineRenderer struct{}

func (cir customInlineRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCustomInline, cir.renderCustom)
}

func (cir customInlineRenderer) renderCustom(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	c := n.(*CustomInline)
	if entering {
		w.WriteByte('<')
		w.WriteString(c.Tag)
		attrs.RenderAll(w, c)
		w.WriteByte('>')
	} else {
		w.WriteString("</")
		w.WriteString(c.Tag)
		w.WriteByte('>')
	}
	return ast.WalkContinue, nil
}

// CustomExt extends markdown with the custom tag renderers.
type CustomExt struct{}

func NewCustomExt() CustomExt {
	return CustomExt{}
}

func (c CustomExt) Extend(m goldmark.Markdown) {
	extenders.AddRenderer(m, customInlineRenderer{}, ord.CustomRenderer)
}
