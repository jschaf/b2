package mdext

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/markdown/attrs"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/ord"
	"github.com/jschaf/bibtex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
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

// FootnoteBody is the block content associated with a footnote link:
//
//   ::: footnote architecture
//   Some *markdown*.
//   :::
type FootnoteBody struct {
	ast.BaseBlock
	Variant FootnoteVariant
	Name    FootnoteName
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
	tag := NewCustomInline("cite")
	attrs.AddClass(tag, "cite-inline")
	txt := ast.NewString([]byte("[" + strconv.Itoa(f.Order) + "]"))
	tag.AppendChild(tag, txt)
	child := f.FirstChild()
	if child != nil && child.Kind() == ast.KindParagraph {
		// Put the cite in the paragraph so it flows in the paragraph.
		child.InsertBefore(child, child.FirstChild(), tag)
	} else {
		// Otherwise, fall back to a separate element.
		f.InsertBefore(f, child, tag)
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
// parsed by colonBlockParser below the location of the corresponding
// FootnoteLink.
type footnoteBodyTransformer struct {
	citeRefsAttacher CitationReferencesAttacher
}

// insertFootnoteBody inserts the body node as the next block node of the most
// distant ancestor from link that's not the document or article.
func insertFootnoteBody(link *FootnoteLink, body *FootnoteBody) {
	ancestor := ast.Node(link)
	for p := ast.Node(link); p.Kind() != KindArticle && p.Kind() != ast.KindDocument; p = p.Parent() {
		ancestor = p
	}
	container := ancestor.Parent()

	// Find the node we should insert the body before.
	before := ancestor.NextSibling()
	for ; before != nil; before = before.NextSibling() {
		if _, ok := before.(*FootnoteBody); !ok {
			break
		}
		if before == body {
			return
		}
	}
	container.InsertBefore(container, before, body)
}

func (fb footnoteBodyTransformer) Transform(doc *ast.Document, source text.Reader, pc parser.Context) {
	links := GetFootnoteLinks(pc)
	bodies := GetFootnoteBodies(pc)
	bibs := GetTOMLMeta(pc).BibPaths
	bibEntries, err := fb.readBibs(bibs)
	if err != nil {
		mdctx.PushError(pc, err)
		return
	}
	refs := NewCitationReferences()
	absPath := GetTOMLMeta(pc).Path
	seenOrders := make(map[string]int)
	order := 1

	// The number of times a key has been used thus far. Useful for
	// generating unique IDs for citations that are used multiple times.
	counts := make(map[FootnoteName]int)

	for _, link := range links {
		// Get the body first.
		var body *FootnoteBody
		if link.Variant == FootnoteVariantCite {
			// Cite bodies are generated from bibtex.
			body = NewFootnoteBody()
		} else {
			// All other variants must have a corresponding body node.
			b, ok := bodies[link.Name]
			if !ok {
				mdctx.PushError(pc, fmt.Errorf("no footnote body for footnote link %q", link.Name))
				continue
			}
			body = b
		}

		// Keep track of already used order numbers. Only citations can be reused,
		// but do it for all variants since it's simpler to keep logic in one spot.
		if o, ok := seenOrders[string(link.Name)]; ok {
			link.Order = o
			body.Order = o
		} else {
			link.Order = order
			body.Order = order
			seenOrders[string(link.Name)] = order
			order++
		}

		// Update ID based on count.
		idSuffix := ""
		if c := counts[link.Name]; c > 0 {
			idSuffix = "-" + strconv.Itoa(c)
		}
		counts[link.Name] += 1

		if link.Variant == FootnoteVariantCite {
			// The citation uses an absolute path because we cut off the references
			// list on the index pages. So instead of broken anchor, deep link to the
			// detail page.
			link.SetAttributeString("data-link-type", "citation")

			// Cite variant bodies are defined by the bibtex entry.
			cr := NewCitationRef()
			c := NewCitation()
			c.Key = bibtex.CiteKey(link.Name)
			bib, ok := bibEntries[c.Key]
			if !ok {
				mdctx.PushError(pc, fmt.Errorf("footnote: no bibtex found for key: %s", c.Key))
				return
			}
			c.Bibtex = bib

			// Render preview for hover when we're not showing side notes. Cite links
			// are preview targets in narrow viewports, when the citation body isn't
			// shown in a side note.
			attrs.AddClass(link, "preview-target")
			b := &bytes.Buffer{}
			citeHTML := bufio.NewWriter(b)
			citeHTML.WriteString("<p>")
			renderCiteRefContent(citeHTML, c)
			citeHTML.WriteString("</p>")
			citeHTML.Flush()
			link.SetAttribute([]byte("data-preview-snippet"), b.Bytes())
			link.SetAttribute([]byte("data-link-type"), LinkCitation)

			link.SetAttributeString("href", absPath+"#"+c.ReferenceID())
			body.Name = link.Name
			body.Variant = link.Variant
			para := ast.NewParagraph()
			para.AppendChild(para, c)
			body.AppendChild(body, para)
			attrs.AddClass(body, "footnote-body-cite")

			cr.Citation = c
			cr.Order = body.Order
			cr.Count = counts[link.Name]
			refs.Refs = append(refs.Refs, cr)
		} else {
			link.SetAttributeString("href", "#footnote-body-"+string(link.Name))
		}

		// Applies to all.
		attrs.AddClass(link, "footnote-link")
		link.SetAttributeString("role", "doc-noteref")
		link.SetAttributeString("id", "footnote-link-"+string(link.Name)+idSuffix)

		attrs.AddClass(body, "footnote-body")
		body.SetAttributeString("id", "footnote-body-"+string(body.Name)+idSuffix)
		body.SetAttributeString("role", "doc-endnote")

		insertFootnoteBody(link, body)
		dist := fb.calcDistance(source, link, pc)
		lineHeight := 18   // from line-height in main.css
		bytesPerLine := 40 // heuristic
		distancePx := (dist/bytesPerLine)*lineHeight + lineHeight
		body.SetAttributeString("style", "margin-top: -"+strconv.Itoa(distancePx)+"px")

		body.addCiteTag() // depends on order
	}

	// Attach the citation references.
	if fb.citeRefsAttacher != nil {
		if err := fb.citeRefsAttacher.Attach(doc, refs); err != nil {
			mdctx.PushError(pc, fmt.Errorf("attach cite references: %w", err))
		}
	}
}

// readBibs returns all bibtex elements from the file paths in bibs merged into
// a map by the key.
func (fb footnoteBodyTransformer) readBibs(bibs []string) (map[bibtex.CiteKey]bibtex.Entry, error) {
	bibEntries := make(map[bibtex.CiteKey]bibtex.Entry)
	for _, bib := range bibs {
		f, err := os.Open(bib)
		if err != nil {
			return nil, fmt.Errorf("citation: read bib file: %w", err)
		}
		entries, err := bibtex.Read(f)
		if err != nil {
			return nil, fmt.Errorf("citation: parse bib file: %w", err)
		}
		for _, elem := range entries {
			bibEntries[elem.Key] = elem
		}
	}
	return bibEntries, nil
}

// calcDistance finds the distance in bytes between the body and the name of
// link in the previous block element (ancestor). Useful to manually position
// footnotes so that they end up closer to the link without using JavaScript.
func (fb footnoteBodyTransformer) calcDistance(source text.Reader, link *FootnoteLink, pc parser.Context) int {
	endPos := 0
	linkPos := 0
	srcBytes := source.Source()

	ancestor := ast.Node(link)
	for p := ast.Node(link); p.Kind() != KindArticle && p.Kind() != ast.KindDocument; p = p.Parent() {
		ancestor = p
	}

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
	w.WriteString(`<a`)
	attrs.RenderAll(w, f)
	w.WriteByte('>')
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
		w.WriteString("<aside")
		attrs.RenderAll(w, f)
		w.WriteByte('>')
	} else {
		w.WriteString("</aside>")
	}
	return ast.WalkContinue, nil
}

// FootnoteExt is the Goldmark extension to render a markdown footnote.
type FootnoteExt struct {
	citeStyle cite.Style
	attacher  CitationReferencesAttacher
}

func NewFootnoteExt(citeStyle cite.Style, attacher CitationReferencesAttacher) *FootnoteExt {
	return &FootnoteExt{citeStyle: citeStyle, attacher: attacher}
}

func (f *FootnoteExt) Extend(m goldmark.Markdown) {
	if f.citeStyle != cite.IEEE {
		panic("unsupported cite style: " + f.citeStyle)
	}
	extenders.AddInlineParser(m, footnoteLinkParser{}, ord.FootnoteLinkParser)
	extenders.AddASTTransform(m, footnoteBodyTransformer{citeRefsAttacher: f.attacher}, ord.FootnoteBodyTransformer)
	extenders.AddRenderer(m, footnoteRenderer{}, ord.FootnoteRenderer)
	extenders.AddRenderer(m, &footnoteIEEERenderer{
		includeRefs: f.attacher != nil,
	}, ord.CitationRenderer)
}
