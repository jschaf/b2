package mdext

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var KindHeader = ast.NewNodeKind("Header")

type Header struct {
	ast.BaseBlock
}

func NewHeader() *Header {
	return &Header{}
}

func (h *Header) Dump(source []byte, level int) {
	ast.DumpHelper(h, source, level, nil, nil)
}

func (h *Header) Kind() ast.NodeKind {
	return KindHeader
}

// headerRenderer is the HTML renderer for a header node.
type headerRenderer struct{}

func NewHeaderRenderer() *headerRenderer {
	return &headerRenderer{}
}

func (hr *headerRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHeader, hr.render)
}

func (hr *headerRenderer) render(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if entering {
		_, _ = w.WriteString("<header>\n")
	} else {
		_, _ = w.WriteString("</header>\n")
	}
	return ast.WalkContinue, nil
}

// headerExt is the Goldmark extension to render a header node.
type headerExt struct{}

func (h *headerExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewHeaderRenderer(), 999)))
}

func NewHeaderExt() *headerExt {
	return &headerExt{}
}
