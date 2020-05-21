package mdext

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindColonBlock = ast.NewNodeKind("ColonBlock")

type ColonBlockName = string

const (
	ColonBlockPreview = "preview"
)

// Preview is a link preview.
type Preview struct {
	// The URL defined as an argument to this colon block, e.g:
	//    ::: preview http://example.com
	URL string
	// The parent colon block that holds the preview content.
	Parent *ColonBlock
}

// ColonBlock parses colon delimited structures inspired by
// https://pandoc.org/MANUAL.html#extension-fenced_divs
// For example:
//
//   ::: preview http://example.com
//   # heading
//   Some *content*
//   :::
type ColonBlock struct {
	ast.BaseBlock

	Name ColonBlockName
	Args string
}

func NewColonBlock() *ColonBlock {
	return &ColonBlock{
		BaseBlock: ast.BaseBlock{},
	}
}

func (c *ColonBlock) Kind() ast.NodeKind {
	return KindColonBlock
}

func (c *ColonBlock) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, nil, nil)
}

// colonBlockParser parsers colon blocks.
type colonBlockParser struct{}

const (
	colonBlockDelim = ":::"
)

func (cbp colonBlockParser) Trigger() []byte {
	return []byte{':'}
}

func (cbp colonBlockParser) Open(_ ast.Node, reader text.Reader, _ parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if !bytes.HasPrefix(line, []byte(colonBlockDelim)) {
		return nil, parser.NoChildren
	}
	i := len(colonBlockDelim)
	reader.Advance(i)
	rest := bytes.Trim(line[i:], " ")
	nameArgs := bytes.SplitN(rest, []byte{' '}, 2)
	cb := NewColonBlock()
	if len(nameArgs) == 1 {
		cb.Name = strings.Trim(string(nameArgs[0]), " ")
	}
	if len(nameArgs) == 2 {
		cb.Args = string(nameArgs[1])
	}
	return cb, parser.HasChildren
}

func (cbp colonBlockParser) Continue(_ ast.Node, reader text.Reader, _ parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if bytes.HasPrefix(line, []byte(colonBlockDelim)) {
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

func (cbp colonBlockParser) Close(node ast.Node, _ text.Reader, pc parser.Context) {
	cb := node.(*ColonBlock)
	switch cb.Name {
	case ColonBlockPreview:
		url := cb.Args
		preview := Preview{
			URL:    url,
			Parent: cb,
		}
		AddPreview(pc, preview)
	}
}

func (cbp colonBlockParser) CanInterruptParagraph() bool {
	return false // No, the colon block must be delimited by a newline.
}

func (cbp colonBlockParser) CanAcceptIndentedLine() bool {
	return false // No, the colon block must not be indented.
}

// colonBlockRenderer renders colon block by omitting them from HTML.
type colonBlockRenderer struct{}

func newColonBlockRenderer() colonBlockRenderer {
	return colonBlockRenderer{}
}

func (cbr colonBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColonBlock, func(util.BufWriter, []byte, ast.Node, bool) (ast.WalkStatus, error) {
		return ast.WalkSkipChildren, nil
	})
}

// ColonBlockExt extends markdown with support for colon blocks, like:
//   ::: preview http://example.com
//   # header
//   :::
type ColonBlockExt struct{}

func NewColonBlockExt() goldmark.Extender {
	return ColonBlockExt{}
}

func (c ColonBlockExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(colonBlockParser{}, 10)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(newColonBlockRenderer(), 1000)))
}
