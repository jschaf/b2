package mdext

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindSmallCaps = ast.NewNodeKind("SmallCaps")

type SmallCaps struct {
	ast.BaseInline
	Segment text.Segment
}

func NewSmallCaps() *SmallCaps {
	return &SmallCaps{}
}

func (s *SmallCaps) Kind() ast.NodeKind {
	return KindSmallCaps
}

func (s *SmallCaps) Dump(source []byte, level int) {
	ast.DumpHelper(s, source, level, nil, nil)
}

type smallCapsParser struct {
}

func (p *smallCapsParser) Trigger() []byte {
	// ' ' indicates any white spaces and a line head
	return []byte{' ', '*', '_', '~', '('}
}

func (p *smallCapsParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	c := line[0]
	consumes := 0
	prev := block.PrecendingCharacter()
	offs := block.LineOffset()
	isEmph := prev == '_' || prev == '*'
	if isEmph && offs >= 2 {
		prevPrev := block.Source()[offs-2]
		// Don't parse intra-word underscores as starters for small caps
		//like FOO_BAR.
		if util.IsAlphaNumeric(prevPrev) {
			return nil
		}
	}
	// advance if current position is not the start of a newline.
	if c == ' ' || c == '*' || c == '_' || c == '~' || c == '(' {
		consumes++
		line = line[1:]
	}
	// We only handle ASCII.
	if len(line) < smallCapsThreshold || line[0] < 'A' || 'Z' < line[0] {
		return nil
	}

	run := 0
	for _, b := range line {
		if 'A' <= b && b <= 'Z' {
			run += 1
		} else {
			break
		}
	}
	if run < smallCapsThreshold {
		return nil
	}
	if run < len(line) {
		next := line[run]
		// Don't use small caps if the upper case chars are followed by anything
		// other than space or punctuation.
		if next != ' ' && next != '\n' && next != '.' && next != '!' &&
			next != '?' && next != ')' && next != '*' && next != ']' {
			return nil
		}
	}
	if consumes != 0 {
		s := segment.WithStop(segment.Start + consumes)
		ast.MergeOrAppendTextSegment(parent, s)
	}

	block.Advance(consumes + run)
	sc := NewSmallCaps()
	sc.Segment = text.NewSegment(
		segment.Start+consumes,
		segment.Start+consumes+run)
	return sc
}

const (
	smallCapsThreshold = 3
)

type smallCapsRenderer struct{}

func NewSmallCapsRenderer() *smallCapsRenderer {
	return &smallCapsRenderer{}
}

func (s *smallCapsRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSmallCaps, s.renderSmallCaps)
}

func (s *smallCapsRenderer) renderSmallCaps(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<span class="small-caps">`)
		sc := node.(*SmallCaps)
		_, _ = w.WriteString(string(sc.Segment.Value(src)))
	} else {
		_, _ = w.WriteString(`</span>`)
	}
	return ast.WalkContinue, nil
}

type SmallCapsExt struct{}

func NewSmallCapsExt() *SmallCapsExt {
	return &SmallCapsExt{}
}

func (sc *SmallCapsExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&smallCapsParser{}, 900)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewSmallCapsRenderer(), 999)))
}
