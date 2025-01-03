package mdext

import (
	"github.com/jschaf/jsc/pkg/markdown/extenders"
	"github.com/jschaf/jsc/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const (
	smallCapsThreshold = 3
)

var KindSmallCaps = ast.NewNodeKind("SmallCaps")

// SmallCaps is an inline node of the text that should be in small caps, e.g.,
// NASA.
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

// smallCapsParser parses text into small caps.
type smallCapsParser struct{}

func (p *smallCapsParser) Trigger() []byte {
	// ' ' indicates whitespace and newlines.
	// We trigger on * so we can parse the small cap inside the emphasized text.
	return []byte{' ', '*', '_', '~', '('}
}

func (p *smallCapsParser) Parse(parent ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()
	c := line[0]
	consumes := 0
	prev := block.PrecendingCharacter()
	offs := block.LineOffset()
	isEmphasis := prev == '_' || prev == '*'
	if isEmphasis && offs >= 2 {
		prevPrev := block.Source()[offs-2]
		// Don't parse intra-word underscores as starters for small caps
		// like FOO_BAR. We don't want FOO_<small-caps>BAR</small-caps>.
		if util.IsAlphaNumeric(prevPrev) {
			return nil
		}
	}
	startChar := byte('\n')
	endChar := byte('\n')
	// Advance if the current position is not the start of a newline.
	if c == ' ' || c == '*' || c == '_' || c == '~' || c == '(' {
		startChar = c
		consumes++
		line = line[1:]
	}
	// We only handle ASCII.
	isCapitalized := 'A' <= line[0] && line[0] <= 'Z'
	isKPrefix := line[0] == 'k' && len(line) > 1 && 'A' <= line[1] && line[1] <= 'Z'
	isCandidate := isCapitalized || isKPrefix
	if len(line) < smallCapsThreshold || !isCandidate {
		return nil
	}

	run := 0
	for _, b := range line {
		if 'A' <= b && b <= 'Z' {
			run += 1
		} else if run == 0 && b == 'k' {
			run += 1 // allow kLOC
		} else {
			break
		}
	}
	// Allow trailing digits as long as we have 2 letters for cases like "SS0".
	if run == smallCapsThreshold-1 && run < len(line) {
		for _, b := range line[run:] {
			if '0' <= b && b <= '9' {
				run += 1
			} else {
				break
			}
		}
	}

	if run < smallCapsThreshold {
		return nil
	}
	if run < len(line) {
		next := line[run]
		endChar = next
		switch next {
		case ' ', '\n', '.', '!', '?', ')', '*', ']', ',':
			// Only use small caps if the run is ended by punctuation or space.
		case 's':
			// s is okay only if followed by punctuation, e.g., TLAs.
			if run+1 < len(line) {
				nextNext := line[run+1]
				switch nextNext {
				case ' ', '\n', '.', '!', '?', ')', '*', ']', ',':
				default:
					return nil
				}
			}
		default:
			return nil
		}
	}

	// We want to convert acronyms inside parens to small caps, e.g.: (NASA).
	// 1 means that parentheses contain the small caps.
	parens := 0
	if startChar == '(' && endChar == ')' {
		parens = 1
	}
	if consumes > 0 && parens == 0 {
		s := segment.WithStop(segment.Start + consumes)
		ast.MergeOrAppendTextSegment(parent, s)
	}

	block.Advance(consumes + run + parens)
	sc := NewSmallCaps()
	sc.Segment = text.NewSegment(
		segment.Start+consumes-parens,
		segment.Start+consumes+run+parens)
	return sc
}

// smallCapsRenderer renders small caps into HTML.
type smallCapsRenderer struct{}

func (s *smallCapsRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSmallCaps, s.renderSmallCaps)
}

func (s *smallCapsRenderer) renderSmallCaps(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<span class="small-caps">`)
		sc := node.(*SmallCaps)
		_, _ = w.Write(sc.Segment.Value(src))
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
	extenders.AddInlineParser(m, &smallCapsParser{}, ord.SmallCapsParser)
	extenders.AddRenderer(m, &smallCapsRenderer{}, ord.SmallCapsRenderer)
}
