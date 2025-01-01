package mdext

import (
	"strconv"

	"github.com/jschaf/jsc/pkg/markdown/attrs"
	"github.com/jschaf/jsc/pkg/markdown/extenders"
	"github.com/jschaf/jsc/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Heading anchor style determines if a paragraph symbol is shown for a heading.
type HeadingAnchorStyle int

const (
	HeadingAnchorStyleNone HeadingAnchorStyle = iota
	HeadingAnchorStyleShow
)

// headingRender writes headings into HTML, customized for my blog.
type headingRender struct {
	style HeadingAnchorStyle
}

func (hr headingRender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, hr.renderHeading)
}

func (hr headingRender) renderHeading(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	h := n.(*ast.Heading)
	const levels = "0123456"
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte(levels[h.Level])
		if h.Attributes() != nil {
			html.RenderAttributes(w, h, html.GlobalAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		switch hr.style {
		case HeadingAnchorStyleNone:
			break
		case HeadingAnchorStyleShow:
			id := attrs.GetStringAttr(h, "id")
			if id == "" {
				break
			}
			_, _ = w.WriteString(`<a class=heading-anchor href="#`)
			_, _ = w.WriteString(id)
			_, _ = w.WriteString(`">¶</a>`)
		default:
			panic("unknown heading anchor style: " + strconv.Itoa(int(hr.style)))
		}
		_, _ = w.WriteString("</h")
		_ = w.WriteByte(levels[h.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

type HeadingExt struct {
	style HeadingAnchorStyle
}

func NewHeadingExt(style HeadingAnchorStyle) HeadingExt {
	return HeadingExt{style: style}
}

func (h HeadingExt) Extend(m goldmark.Markdown) {
	extenders.AddRenderer(m, headingRender{style: h.style}, ord.HeadingRenderer)
}
