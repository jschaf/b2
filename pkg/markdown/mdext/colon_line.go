package mdext

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindColonLine represents a block like:
//   :toc: right
var KindColonLine = ast.NewNodeKind("ColonLine")

type ColonLineName string

const (
	ColonLineTOC ColonLineName = "toc"
)

// ColonLine parses colon delimited directives inspired by
// https://asciidoctor.org/docs/asciidoc-syntax-quick-reference/#table-of-contents-toc
// For example to create a right-aligned TOC:
//
//   :toc: right
type ColonLine struct {
	ast.BaseBlock
	Name ColonLineName
	Args string
}

func NewColonLine() *ColonLine {
	return &ColonLine{}
}

func (c *ColonLine) Kind() ast.NodeKind {
	return KindColonLine
}

func (c *ColonLine) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, nil, nil)
}

// ColonLineParser parsers colon blocks.
type ColonLineParser struct{}

func (clp ColonLineParser) Trigger() []byte {
	return []byte{':'}
}

func (clp ColonLineParser) Open(_ ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	const minLen = len(":toc:")
	if len(line) < minLen || line[0] != ':' {
		return nil, parser.Continue
	}

	// Consume the word in the colons.
	i := 1
	for ; i < len(line); i++ {
		if !('a' <= line[i] && line[i] <= 'z') && line[i] != '_' {
			break
		}
	}

	if i >= len(line) || line[i] != ':' {
		return nil, parser.Continue
	}
	i++ // consume closing colon

	reader.AdvanceLine()
	cb := NewColonLine()
	cb.Name = ColonLineName(line[1 : i-1])
	cb.Args = ""
	if i < len(line) {
		cb.Args = string(bytes.TrimSpace(line[i:]))
	}
	switch cb.Name {
	case ColonLineTOC:
		toc := NewTOC()
		SetTOC(pc, toc)
		return toc, parser.Close
	}
	return cb, parser.Close
}

func (clp ColonLineParser) Continue(_ ast.Node, _ text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

func (clp ColonLineParser) Close(_ ast.Node, _ text.Reader, _ parser.Context) {
}

func (clp ColonLineParser) CanInterruptParagraph() bool {
	return false // No, the colon block must be delimited by a newline.
}

func (clp ColonLineParser) CanAcceptIndentedLine() bool {
	return false // No, the colon block must not be indented.
}

// ColonLineRenderer renders colon block by omitting them from HTML.
type ColonLineRenderer struct{}

func newColonLineRenderer() ColonLineRenderer {
	return ColonLineRenderer{}
}

func (cbr ColonLineRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColonLine, func(util.BufWriter, []byte, ast.Node, bool) (ast.WalkStatus, error) {
		return ast.WalkSkipChildren, nil
	})
}

// ColonLineExt extends markdown with support for colon blocks, like:
//   ::: preview http://example.com
//   # header
//   :::
type ColonLineExt struct{}

func NewColonLineExt() goldmark.Extender {
	return ColonLineExt{}
}

func (c ColonLineExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(ColonLineParser{}, 10)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(newColonLineRenderer(), 1000)))
}
