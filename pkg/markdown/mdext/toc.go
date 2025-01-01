package mdext

import (
	"bytes"
	"strconv"

	"github.com/jschaf/jsc/pkg/markdown/asts"
	"github.com/jschaf/jsc/pkg/markdown/attrs"
	"github.com/jschaf/jsc/pkg/markdown/extenders"
	"github.com/jschaf/jsc/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindTOC represents a TOC node.
var KindTOC = ast.NewNodeKind("TOC")

type TOCStyle int

const (
	// The inclusive start of TOC headings. 2 means consider H2 as the top level
	// TOC.
	tocStartLevel = 2
)

const (
	TOCStyleNone TOCStyle = iota
	TOCStyleShow
)

// TOC contains directives to format a table of contents.
// TOC nodes are created from the ColonLine parser.
type TOC struct {
	ast.BaseBlock
	Headings []*ast.Heading
}

func NewTOC() *TOC {
	return &TOC{}
}

func (c *TOC) Kind() ast.NodeKind {
	return KindTOC
}

func (c *TOC) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, nil, nil)
}

var tocCtxKey = parser.NewContextKey()

func GetTOC(pc parser.Context) (*TOC, bool) {
	r := pc.Get(tocCtxKey)
	if r == nil {
		return nil, false
	}
	return r.(*TOC), true
}

func SetTOC(pc parser.Context, toc *TOC) {
	pc.Set(tocCtxKey, toc)
}

// tocTransformer adds heading entries to the TOC node.
type tocTransformer struct {
	// How many heading levels (1-based) to include in the TOC. If depth is 0,
	// defaults to 3. For example, a depth of 2 includes H1 and H2 headings in the
	// TOC.
	depth int
	style TOCStyle
}

func newTOCTransformer(s TOCStyle) tocTransformer {
	return tocTransformer{style: s}
}

func (t tocTransformer) Transform(node *ast.Document, _ text.Reader, pc parser.Context) {
	if t.style == TOCStyleNone {
		return
	}
	toc, ok := GetTOC(pc)
	if !ok {
		return
	}

	depth := t.depth
	if t.depth == 0 {
		depth = 4
	}

	headings := make([]*ast.Heading, 0, 3*depth) // assume 3 headings per level
	_ = asts.WalkHeadings(node, func(h *ast.Heading) (ast.WalkStatus, error) {
		// We ignore H1 because that's the title.
		if h.Level > 1 && h.Level <= depth {
			headings = append(headings, h)
		}
		return ast.WalkSkipChildren, nil
	})

	// The count per 1-indexed ast.Heading.Level: 1-6. Must increment before use.
	counts := make([]int, 7)
	l, _ := createTOCListLevel(headings, 2, counts) // start at 2 since 1 is the title
	toc.AppendChild(toc, l)
}

// createTOCListLevel creates a list containing a single level of headings and
// recurses to create deeper levels.
func createTOCListLevel(headings []*ast.Heading, level int, counts []int) (*ast.List, int) {
	l := ast.NewList('.') // period is a marker for an ordered list
	attrs.AddClass(l, "toc-list", "toc-level-"+strconv.Itoa(level))
	l.Start = 1
	i := 0
	for i < len(headings) {
		h := headings[i]
		switch {
		case h.Level < level:
			// Reset counts because numbering starts over with a new parent heading.
			resetCounts(counts, level)
			// Let the parent createTOCListLevel handle this heading.
			return l, i
		case h.Level == level:
			li := ast.NewListItem(i)
			prefix := buildCountTag(counts, level)
			li.AppendChild(li, prefix)

			id := attrs.GetStringAttr(h, "id")
			if id != "" {
				link := ast.NewLink()
				link.Destination = []byte("#" + id)
				asts.Reparent(link, CloneNode(h)) // clone to avoid moving actual headings
				li.AppendChild(li, link)
			} else {
				asts.Reparent(li, CloneNode(h)) // clone to avoid moving actual headings
			}
			l.AppendChild(l, li)
			i++
			continue
		case h.Level > level:
			li := ast.NewListItem(i)
			childL, n := createTOCListLevel(headings[i:], level+1, counts)
			li.AppendChild(li, childL)
			l.AppendChild(l, li)
			i += n
			continue
		default:
			panic("unreachable")
		}
	}
	return l, i
}

func resetCounts(counts []int, level int) {
	for i := level; i < len(counts); i++ {
		counts[i] = 0
	}
}

// buildCountTag builds an HTML string representing the current TOC level at
// depth of level. For example, a level of 3 produces an HTML tag with contents
// like "4.2" which means this TOC entry is the 4th <h2> and 2nd <h3>. We ignore
// levels less than tocStartLevel which is why there's no <h1> in the example.
func buildCountTag(counts []int, level int) *ast.String {
	counts[level]++
	b := new(bytes.Buffer)
	b.WriteString("<span class=toc-ordering>")
	for i := tocStartLevel; i <= level; i++ {
		b.WriteString(strconv.Itoa(counts[i]))
		if i < level {
			b.WriteByte('.')
		}
	}
	b.WriteString("</span>")
	b.WriteByte(' ')
	s := ast.NewString(b.Bytes())
	s.SetCode(true)
	return s
}

// tocRenderer is the HTML renderer for a TOC node.
type tocRenderer struct {
	style TOCStyle
}

func newTOCRenderer(s TOCStyle) tocRenderer {
	return tocRenderer{style: s}
}

func (tr tocRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	switch tr.style {
	case TOCStyleNone:
		reg.Register(KindTOC, renderTOCNop)
	case TOCStyleShow:
		reg.Register(KindTOC, renderTOC)
	default:
		panic("unknown TOC style: " + strconv.Itoa(int(tr.style)))
	}
}

func renderTOC(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<div class=toc>")
	} else {
		w.WriteString("</div>")
	}
	return ast.WalkContinue, nil
}

func renderTOCNop(_ util.BufWriter, _ []byte, _ ast.Node, _ bool) (ast.WalkStatus, error) {
	return ast.WalkSkipChildren, nil
}

type TOCExt struct {
	style TOCStyle
}

func NewTOCExt(s TOCStyle) goldmark.Extender {
	return TOCExt{style: s}
}

func (t TOCExt) Extend(m goldmark.Markdown) {
	extenders.AddASTTransform(m, newTOCTransformer(t.style), ord.TOCTransformer)
	extenders.AddRenderer(m, newTOCRenderer(t.style), ord.TOCRenderer)
}
