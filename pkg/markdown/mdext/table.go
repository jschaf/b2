package mdext

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jschaf/jsc/pkg/markdown/asts"
	"github.com/jschaf/jsc/pkg/markdown/extenders"
	"github.com/jschaf/jsc/pkg/markdown/mdctx"
	"github.com/jschaf/jsc/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	ast2 "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindTableCaption = ast.NewNodeKind("TableCaption")

type TableCaption struct {
	ast.BaseInline
	Order int
}

func NewTableCaption() *TableCaption {
	return &TableCaption{}
}

func (t *TableCaption) Kind() ast.NodeKind {
	return KindTableCaption
}

func (t *TableCaption) Dump(source []byte, level int) {
	ast.DumpHelper(t, source, level, nil, nil)
}

const tableCaptionMarker = "TABLE:"

var tableNumCtxKey = parser.NewContextKey()

// getNextTableNum gets the next table number to use in the table caption like
// "Table 1: some caption".
func getNextTableNum(pc parser.Context) int {
	num := pc.Get(tableNumCtxKey)
	n := 0
	if _, ok := num.(int); ok {
		n = num.(int)
	}
	n++
	pc.Set(tableNumCtxKey, n)
	return n
}

// tableCaptionTransformer is an AST transformer that moves paragraphs like
// "TABLE: foo" as a <caption> nested under the following table.
type tableCaptionTransformer struct{}

func (t tableCaptionTransformer) Transform(doc *ast.Document, r text.Reader, pc parser.Context) {
	err := asts.WalkKind(ast2.KindTable, doc, func(n ast.Node) (ast.WalkStatus, error) {
		t := n.(*ast2.Table)
		capt := t.PreviousSibling()
		if !isTableCaption(capt, r) {
			return ast.WalkSkipChildren, nil
		}

		// Trim the marker "TABLE:"
		txt := capt.FirstChild().(*ast.Text)
		txt.Segment.Start += len(tableCaptionMarker)
		tblCapt := NewTableCaption()
		tblCapt.Order = getNextTableNum(pc)
		asts.Reparent(tblCapt, capt)
		// Remove the old caption which is empty because we moved (reparent).
		parent := capt.Parent()
		parent.RemoveChild(parent, capt)
		// Make the caption the first elem.
		t.InsertBefore(t, t.FirstChild(), tblCapt)

		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		mdctx.PushError(pc, fmt.Errorf("table caption transform walk: %w", err))
	}
}

func isTableCaption(n ast.Node, r text.Reader) bool {
	if n == nil || n.Kind() != ast.KindParagraph || n.FirstChild() == nil ||
		n.FirstChild().Kind() != ast.KindText {
		return false
	}
	s := string(n.FirstChild().Text(r.Source()))
	return strings.HasPrefix(s, tableCaptionMarker)
}

type tableCaptionRenderer struct{}

func (t tableCaptionRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindTableCaption, t.renderTableCaption)
}

func (t tableCaptionRenderer) renderTableCaption(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	tc := n.(*TableCaption)
	if entering {
		w.WriteString("<caption>")
		w.WriteString("<span class=table-caption-order>Table ")
		w.WriteString(strconv.Itoa(tc.Order))
		w.WriteString(": </span>")
	} else {
		w.WriteString("</caption>")
	}
	return ast.WalkContinue, nil
}

type TableExt struct{}

func NewTableExt() TableExt {
	return TableExt{}
}

func (t TableExt) Extend(m goldmark.Markdown) {
	extenders.AddParaTransform(m, extension.NewTableParagraphTransformer(), ord.TableParaTransformer)
	extenders.AddASTTransform(m, tableCaptionTransformer{}, ord.TableCaptionTransformer)
	extenders.AddRenderer(m, extension.NewTableHTMLRenderer(), ord.TableRenderer)
	extenders.AddRenderer(m, tableCaptionRenderer{}, ord.TableCaptionRenderer)
}
