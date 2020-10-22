package mdext

import (
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/ord"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindColonBlock = ast.NewNodeKind("ColonBlock")

type ColonBlockName string

const (
	ColonBlockPreview  ColonBlockName = "preview"
	ColonBlockSideNote ColonBlockName = "side-note"
)

// Preview is a link preview.
type Preview struct {
	// The URL defined as an argument to this colon block, e.g:
	//    ::: preview http://example.com
	URL string
	// The parent colon block that holds the preview content.
	Parent *ColonBlock
}

var previewCtxKey = parser.NewContextKey()

// AddPreview adds a preview to the context so that it can be rendered into
// the corresponding link.
func AddPreview(pc parser.Context, p Preview) {
	if existing := pc.Get(previewCtxKey); existing == nil {
		pc.Set(previewCtxKey, make(map[string]Preview))
	}

	previews := pc.Get(previewCtxKey).(map[string]Preview)
	previews[p.URL] = p
}

// GetPreview returns the preview, if any, for the URL. Returns an empty Preview
// and false if no preview exists for the URL.
func GetPreview(pc parser.Context, url string) (Preview, bool) {
	previews, ok := pc.Get(previewCtxKey).(map[string]Preview)
	if !ok {
		return Preview{}, false
	}
	p, ok := previews[url]
	return p, ok
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
	reader.AdvanceLine()
	rest := bytes.Trim(line[len(colonBlockDelim):], " \t\n")
	nameArgs := bytes.SplitN(rest, []byte{' '}, 2)
	cb := NewColonBlock()
	if len(nameArgs) >= 1 {
		cb.Name = ColonBlockName(strings.Trim(string(nameArgs[0]), " "))
	}
	if len(nameArgs) == 2 {
		cb.Args = strings.Trim(string(nameArgs[1]), " \n")
	}
	return cb, parser.HasChildren
}

func (cbp colonBlockParser) Continue(_ ast.Node, reader text.Reader, _ parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if bytes.HasPrefix(line, []byte(colonBlockDelim)) {
		reader.AdvanceLine()
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
	reg.Register(KindColonBlock, cbr.renderColonBlock)
}
func (cbr colonBlockRenderer) renderColonBlock(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	c := n.(*ColonBlock)
	switch c.Name {
	case ColonBlockPreview:
		return ast.WalkSkipChildren, nil
	case ColonBlockSideNote:
		if entering {
			w.WriteString(`<aside class="side-note" id="sn-`)
			w.WriteString(c.Args)
			w.WriteString(`">`)
		} else {
			w.WriteString("</aside>")
		}
		return ast.WalkContinue, nil
	default:
		return ast.WalkContinue, fmt.Errorf("render unknown colon block name %q", c.Name)
	}
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
	extenders.AddBlockParser(m, colonBlockParser{}, ord.ColonBlockParser)
	extenders.AddRenderer(m, newColonBlockRenderer(), ord.ColonBlockRenderer)
}
