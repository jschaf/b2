package mdext

import (
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var KindTime = ast.NewNodeKind("Time")

type Time struct {
	ast.BaseInline
	date time.Time
}

func (t *Time) Dump(source []byte, level int) {
	ast.DumpHelper(t, source, level, nil, nil)
}

func (t *Time) Kind() ast.NodeKind {
	return KindTime
}

func NewTime(date time.Time) *Time {
	return &Time{
		date: date,
	}
}

type TimeRenderer struct {
	html.Config
}

func (t *TimeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindTime, t.renderTime)
}

func (t *TimeRenderer) renderTime(w util.BufWriter, source []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if entering {
		n := node.(*Time)
		_, _ = w.WriteString("\n<time datetime=\"")
		_, _ = w.WriteString(n.date.UTC().Format("2006-01-02"))
		_, _ = w.WriteString("\">")
		_, _ = w.WriteString(n.date.Format("January _2, 2006"))
		_, _ = w.WriteString("</time>\n")
	}
	return ast.WalkContinue, nil
}

func NewTimeHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &TimeRenderer{Config: html.NewConfig()}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

type timeExt struct {
}

func NewTimeExt() *timeExt {
	return &timeExt{}
}

func (t *timeExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewTimeHTMLRenderer(), 500)))
}
