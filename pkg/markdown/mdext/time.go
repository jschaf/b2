package mdext

import (
	"time"

	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/ord"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var KindTime = ast.NewNodeKind("Time")

// Time is the date time from the TOML metadata of the publish date of a post.
type Time struct {
	ast.BaseInline
	Date time.Time
}

func NewTime(date time.Time) *Time {
	return &Time{
		Date: date,
	}
}

func (t *Time) Dump(source []byte, level int) {
	ast.DumpHelper(t, source, level, nil, nil)
}

func (t *Time) Kind() ast.NodeKind {
	return KindTime
}

// timeRenderer renders HTML for a Time node.
type timeRenderer struct {
	html.Config
}

func newTimeRenderer() renderer.NodeRenderer {
	r := &timeRenderer{Config: html.NewConfig()}
	return r
}

func (t *timeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindTime, t.renderTime)
}

func (t *timeRenderer) renderTime(w util.BufWriter, _ []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if entering {
		n := node.(*Time)
		_, _ = w.WriteString("\n<time datetime=\"")
		_, _ = w.WriteString(n.Date.UTC().Format("2006-01-02"))
		_, _ = w.WriteString("\">")
		_, _ = w.WriteString(n.Date.Format("January _2, 2006"))
		_, _ = w.WriteString("</time>\n")
	}
	return ast.WalkContinue, nil
}

// TimeExt extends the goldmark markdown renderer to support a time node.
type TimeExt struct{}

func NewTimeExt() *TimeExt {
	return &TimeExt{}
}

func (t *TimeExt) Extend(m goldmark.Markdown) {
	extenders.AddRenderer(m, newTimeRenderer(), ord.TimeRenderer)
}
