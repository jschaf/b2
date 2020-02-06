// This extension parses YAML metadata blocks and stores metadata in
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

// PostMeta is the TOML metadata of a post.
type PostMeta struct {
	Slug string
	Date time.Time
}

type data struct {
	Map   PostMeta
	Error error
	Node  ast.Node
}

var contextKey = parser.NewContextKey()

// GetTOMLMeta returns a TOML metadata.
func GetTOMLMeta(pc parser.Context) PostMeta {
	v := pc.Get(contextKey)
	if v == nil {
		return PostMeta{}
	}
	d := v.(*data)
	return d.Map
}

type tomlMeta struct {
}

var defaultMetaParser = &tomlMeta{}

// NewParser returns a BlockParser that can parse YAML metadata blocks.
func NewParser() parser.BlockParser {
	return defaultMetaParser
}

func isSeparator(line []byte) bool {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	for i := 0; i < len(line); i++ {
		if line[i] != '+' {
			return false
		}
	}
	return true
}

func (b *tomlMeta) Trigger() []byte {
	return []byte{'+'}
}

func (b *tomlMeta) Open(_ ast.Node, reader text.Reader, _ parser.Context) (ast.Node, parser.State) {
	lineNum, _ := reader.Position()
	if lineNum != 0 {
		return nil, parser.NoChildren
	}
	line, _ := reader.PeekLine()
	if isSeparator(line) {
		return ast.NewTextBlock(), parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (b *tomlMeta) Continue(node ast.Node, reader text.Reader, _ parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if isSeparator(line) {
		reader.Advance(segment.Len())
		return parser.Close
	}
	node.Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (b *tomlMeta) Close(node ast.Node, reader text.Reader, pc parser.Context) {
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

	pc.Set(contextKey, d)

	if d.Error == nil {
		node.Parent().RemoveChild(node.Parent(), node)
	}
}

func (b *tomlMeta) CanInterruptParagraph() bool {
	return false
}

func (b *tomlMeta) CanAcceptIndentedLine() bool {
	return false
}

type tomlFront struct {
	Table bool
}

// TOMLFrontmatter is a extension for the goldmark.
var TOMLFrontmatter = &tomlFront{}

// New returns a new TOMLFrontmatter extension.
func New() goldmark.Extender {
	return &tomlFront{}
}

func (e *tomlFront) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewParser(), 0),
		),
	)
}
