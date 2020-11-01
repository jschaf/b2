package mdext

import (
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/markdown/attrs"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"strconv"
	"strings"
)

type FootnoteName string

type FootnoteVariant string

const (
	FootnoteVariantCite   FootnoteVariant = "cite"
	FootnoteVariantSide   FootnoteVariant = "side"
	FootnoteVariantMargin FootnoteVariant = "margin"
	FootnoteVariantPara   FootnoteVariant = "para"
)

var (
	KindFootnoteLink = ast.NewNodeKind("FootnoteLink")
	KindFootnoteBody = ast.NewNodeKind("FootnoteBody")
)

// A FootnoteLink marks the location that the FootnoteBody describes.
//
// [^side:arch] or [^para:arch] or [^margin:arch] or [^@spanner2012]
type FootnoteLink struct {
	ast.BaseInline
	Variant FootnoteVariant
	// The full name including the type prefix, like "side:arch". The name for
	// cite variants doesn't include the "@" prefix.
	Name FootnoteName
	// The order of this footnote in the document. Only applies to the sidenote
	// and cite variant. Always 0 for other variants. Duplicate citations re-use
	// the earliest order number.
	Order int
}

func NewFootnoteLink() *FootnoteLink {
	return &FootnoteLink{}
}

func (f *FootnoteLink) Kind() ast.NodeKind {
	return KindFootnoteLink
}

func (f *FootnoteLink) Dump(source []byte, level int) {
	ast.DumpHelper(f, source, level, nil, nil)
}

func (f *FootnoteLink) FootnoteOrder(nextOrder int, seen map[string]int, pc parser.Context) (FnAction, string) {
	if f.Variant != FootnoteVariantSide {
		return FnOrderKeep, string(f.Name)
	}
	if _, ok := seen[string(f.Name)]; ok {
		panic(fmt.Sprintf("footnote %q already seen but keys should be unique", string(f.Name)))
	}
	f.Order = nextOrder
	bodies := GetFootnoteBodies(pc)
	body, ok := bodies[f.Name]
	if !ok {
		panic(fmt.Sprintf("no footnote body found for footnote %q", f.Name))
	}
	body.Order = nextOrder
	body.addCiteTag()
	return FnOrderNext, string(f.Name)
}

// FootnoteBody is the block content associated with a footnote link:
//
//   ::: footnote architecture
//   Some *markdown*.
//   :::
type FootnoteBody struct {
	ast.BaseBlock
	Variant FootnoteVariant
	Name    FootnoteName
	// The distance in bytes between the link and this body. The link is always
	// above the body because it's reordered in an AST transformer. Helpful to
	// render the body close to the link.
	LinkDistance int
	// The order of this footnote in the document. Only applies to the sidenote
	// and cite variant. Always 0 for other variants. Duplicate citations re-use
	// the earliest order number.
	Order int
}

func NewFootnoteBody() *FootnoteBody {
	return &FootnoteBody{}
}

func (f *FootnoteBody) Kind() ast.NodeKind {
	return KindFootnoteBody
}

func (f *FootnoteBody) Dump(source []byte, level int) {
	ast.DumpHelper(f, source, level, nil, nil)
}

func (f *FootnoteBody) addCiteTag() {
	cite := NewCustomInline("cite")
	attrs.AddClass(cite, "cite-inline")
	txt := ast.NewString([]byte("[" + strconv.Itoa(f.Order) + "]"))
	cite.AppendChild(cite, txt)
	child := f.FirstChild()
	if child != nil && child.Kind() == ast.KindParagraph {
		// Put the cite in the paragraph so it flows in the paragraph.
		child.InsertBefore(child, child.FirstChild(), cite)
	} else {
		// Otherwise, fall back to a separate element.
		f.InsertBefore(f, child, cite)
	}
}

// footnoteLinkParser is an inline parser to parse footnote links like
// [^side:foo], or [^margin:qux].
type footnoteLinkParser struct{}

func (f footnoteLinkParser) Trigger() []byte {
	return []byte{'['}
}

func parseFootnoteName(s string) (FootnoteName, FootnoteVariant, error) {
	switch {
	case strings.HasPrefix(s, string(FootnoteVariantSide)+":"):
		return FootnoteName(s), FootnoteVariantSide, nil
	case strings.HasPrefix(s, string(FootnoteVariantMargin)+":"):
		return FootnoteName(s), FootnoteVariantMargin, nil
	case strings.HasPrefix(s, string(FootnoteVariantPara)+":"):
		return FootnoteName(s), FootnoteVariantPara, nil
	case strings.HasPrefix(s, "@"):
		return FootnoteName(s[1:]), FootnoteVariantCite, nil // drop @ from cite key
	default:
		return "", "", fmt.Errorf("unknown footnote variant: %q", s)
	}
}

func (f footnoteLinkParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	pos := 1
	if pos >= len(line) || line[pos] != '^' {
		return nil
	}
	pos++ // consume '^'
	if pos >= len(line) {
		return nil
	}
	open := pos
	closure := util.FindClosure(line[pos:], '[', ']', false, false)
	if closure < 0 {
		return nil
	}
	closes := pos + closure
	value := string(block.Value(text.NewSegment(segment.Start+open, segment.Start+closes)))

	block.Advance(closes + 1)
	link := NewFootnoteLink()
	name, variant, err := parseFootnoteName(value)
	if err != nil {
		mdctx.PushError(pc, fmt.Errorf("parse inline footnote: %w", err))
		return nil
	}
	link.Name = name
	link.Variant = variant
	AddFootnoteLink(pc, link)
	return link
}

// footnoteBodyTransformer adds FootnoteBody nodes stored in parser.Context
// parsed by colonBlockParser below the location of the FootnoteLink.
type footnoteBodyTransformer struct {
}

// farthestBlockAncestor returns the farthest ancestor node that's not the
// document or article. This is useful to figure out where to put footnote
// bodies so that the body is the next block element after its footnote link.
func farthestBlockAncestor(node ast.Node) ast.Node {
	parent := node
	for p := node; p.Kind() != KindArticle && p.Kind() != ast.KindDocument; p = p.Parent() {
		parent = p
	}
	return parent
}

func (fb footnoteBodyTransformer) Transform(_ *ast.Document, source text.Reader, pc parser.Context) {
	links := GetFootnoteLinks(pc)
	bodies := GetFootnoteBodies(pc)
	for _, link := range links {
		body, ok := bodies[link.Name]
		if !ok {
			mdctx.PushError(pc, fmt.Errorf("no footnote body for footnote link %q", link.Name))
			continue
		}
		// Place the footnote body immediately after the block containing the
		// corresponding footnote link.
		ancestor := farthestBlockAncestor(link)
		if b, ok := ancestor.NextSibling().(*FootnoteBody); !ok || b != body {
			// Only move the body if it's not already in the correct spot.
			container := ancestor.Parent()
			container.InsertAfter(container, ancestor, body)
		}
		body.LinkDistance = fb.calcDistance(source, ancestor, link, pc)
	}
}

// calcDistance finds the distance in bytes between the body and the name of
// link in the previous block element (ancestor). Useful to manually position
// footnotes so that they end up closer to the link without using JavaScript.
func (fb footnoteBodyTransformer) calcDistance(source text.Reader, ancestor ast.Node, link *FootnoteLink, pc parser.Context) int {
	endPos := 0
	linkPos := 0
	srcBytes := source.Source()
	err := ast.Walk(ancestor, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		// ast.TextBlock is normally the only node type with RawText that we can use
		// to get segment ends and to find the footnote name.
		if n.Type() != ast.TypeBlock {
			return ast.WalkSkipChildren, nil
		}
		lines := n.Lines()
		// Update the endPos.
		if lines.Len() > 0 {
			endPos = lines.At(lines.Len() - 1).Stop
		}
		// Find the position of the name in the ancestor.
		for _, segment := range lines.Sliced(0, lines.Len()) {
			bs := srcBytes[segment.Start:segment.Stop]
			pos := bytes.Index(bs, []byte(link.Name))
			if pos >= 0 {
				linkPos = segment.Start + pos
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		mdctx.PushError(pc, fmt.Errorf("calc distance for footnote body offset: %w", err))
		return 0
	}
	return endPos - linkPos
}

// footnoteRenderer renders both FootnoteLink and FootnoteBody.
type footnoteRenderer struct {
}

func (fr footnoteRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFootnoteLink, fr.renderFootnoteLink)
	reg.Register(KindFootnoteBody, fr.renderFootnoteBody)
}

func (fr footnoteRenderer) renderFootnoteLink(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	f := n.(*FootnoteLink)
	w.WriteString(`<a class="footnote-link" role="doc-noteref" href="#footnote-body-`)
	w.WriteString(string(f.Name))
	w.WriteString(`" id="footnote-link-`)
	w.WriteString(string(f.Name))
	w.WriteString(`">`)
	switch f.Variant {
	case FootnoteVariantPara: // no indicator for a paragraph note
	case FootnoteVariantMargin: // no indicator for a margin note
	case FootnoteVariantSide, FootnoteVariantCite:
		w.WriteString("<cite>[")
		w.WriteString(strconv.Itoa(f.Order))
		w.WriteString("]</cite>")
	default:
		return ast.WalkStop, fmt.Errorf("unknown footnote variant %q in renderFootnoteLink", f.Variant)
	}
	w.WriteString(`</a>`)
	return ast.WalkSkipChildren, nil
}

func (fr footnoteRenderer) renderFootnoteBody(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	f := n.(*FootnoteBody)
	if entering {
		w.WriteString(`<aside class="footnote-body" id="footnote-body-`)
		w.WriteString(string(f.Name))
		w.WriteString(`" role="doc-endnote"`)
		w.WriteString(` style="margin-top: -`)
		lineHeight := 18
		bytesPerLine := 40 // heuristic
		distancePx := (f.LinkDistance/bytesPerLine)*lineHeight + lineHeight
		w.WriteString(strconv.Itoa(distancePx))
		w.WriteString(`px">`)
	} else {
		w.WriteString("</aside>")
	}
	return ast.WalkContinue, nil
}

// FootnoteExt is the Goldmark extension to render a markdown footnote.
type FootnoteExt struct{}

func NewFootnoteExt() *FootnoteExt {
	return &FootnoteExt{}
}

func (f *FootnoteExt) Extend(m goldmark.Markdown) {
	extenders.AddInlineParser(m, footnoteLinkParser{}, ord.FootnoteLinkParser)
	extenders.AddASTTransform(m, footnoteBodyTransformer{}, ord.FootnoteBodyTransformer)
	extenders.AddASTTransform(m, footnoteOrderTransformer{}, ord.FootnoteOrderTransformer)
	extenders.AddRenderer(m, footnoteRenderer{}, ord.FootnoteRenderer)
}
