// This extension parses TOML metadata blocks and stores metadata in
// parser.Context.
package mdext

import (
	"bytes"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const tomlSep = '+'

// PostMeta is the TOML metadata of a post.
type PostMeta struct {
	Slug  string
	Title string
	Date  time.Time
}

type data struct {
	Map   PostMeta
	Error error
	Node  ast.Node
}

var ctxKey = parser.NewContextKey()

// GetTOMLMeta returns a TOML metadata.
func GetTOMLMeta(pc parser.Context) PostMeta {
	v := pc.Get(ctxKey)
	if v == nil {
		return PostMeta{}
	}
	d := v.(*data)
	return d.Map
}

type tomlMeta struct {
}

var defaultTOMLMetaParser = &tomlMeta{}

// NewTOMLParser returns a BlockParser that can parse TOML metadata blocks.
func NewTOMLParser() parser.BlockParser {
	return defaultTOMLMetaParser
}

func isTOMLSep(line []byte) bool {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	for i := 0; i < len(line); i++ {
		if line[i] != tomlSep {
			return false
		}
	}
	return true
}

func (t *tomlMeta) Trigger() []byte {
	return []byte{tomlSep}
}

func (t *tomlMeta) Open(_ ast.Node, reader text.Reader, _ parser.Context) (ast.Node, parser.State) {
	lineNum, _ := reader.Position()
	if lineNum != 0 {
		return nil, parser.NoChildren
	}
	line, _ := reader.PeekLine()
	if isTOMLSep(line) {
		return ast.NewTextBlock(), parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (t *tomlMeta) Continue(node ast.Node, reader text.Reader, _ parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if isTOMLSep(line) {
		reader.Advance(segment.Len())
		return parser.Close
	}
	node.Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (t *tomlMeta) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	lines := node.Lines()
	var buf bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		buf.Write(segment.Value(reader.Source()))
	}
	d := &data{}
	d.Node = node
	meta := &PostMeta{}
	if err := toml.Unmarshal(buf.Bytes(), &meta); err != nil {
		d.Error = err
	} else {
		d.Map = *meta
	}

	pc.Set(ctxKey, d)

	if d.Error == nil {
		node.Parent().RemoveChild(node.Parent(), node)
	}
}

func (t *tomlMeta) CanInterruptParagraph() bool {
	return false
}

func (t *tomlMeta) CanAcceptIndentedLine() bool {
	return false
}

type tomlFront struct {
	Table bool
}

// New returns a new TOMLFrontmatter extension.
func NewTOMLExt() goldmark.Extender {
	return &tomlFront{}
}

func (t *tomlFront) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewTOMLParser(), 0),
		),
	)
}
