package mdext

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindCitation = ast.NewNodeKind("citation")

// Citation is an inline node representing a citation.
type Citation struct {
	ast.BaseInline
	ID     string
	Prefix string
	Suffix string
}

var citeBottom = parser.NewContextKey()

type citationParser struct {
}

func (c citationParser) Trigger() []byte {
	return []byte{'[', ']'}
}

func (c citationParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	switch line[0] {
	case '[':
		pc.Set(citeBottom, pc.LastDelimiter())

	case ']':

	}
	return nil
}

// citationParagraphTransformer promotes
type citationParagraphTransformer struct{}

func (c citationParagraphTransformer) Transform(node *ast.Paragraph, reader text.Reader, pc parser.Context) {
	fmt.Println("\ncitationParagraphTransformer(before):")
	node.Dump(reader.Source(), 0)
	defer func() {
		var root ast.Node
		root = node
		for root.Parent() != nil {
			root = root.Parent()
		}
		fmt.Println("\ncitationParagraphTransformer(after):")
		root.Dump(reader.Source(), 0)
	}()
	lines := node.Lines()
	block := text.NewBlockReader(reader.Source(), lines)
	removes := [][2]int{}
	for {
		start, end := parseCitation(block, pc)
		if start > -1 {
			if start == end {
				end++
			}
			removes = append(removes, [2]int{start, end})
			continue
		}
		break
	}

	offset := 0
	for _, remove := range removes {
		if lines.Len() == 0 {
			break
		}
		s := lines.Sliced(remove[1]-offset, lines.Len())
		lines.SetSliced(0, remove[0]-offset)
		lines.AppendAll(s)
		offset = remove[1]
	}

	if lines.Len() == 0 {
		t := ast.NewTextBlock()
		t.SetBlankPreviousLines(node.HasBlankPreviousLines())
		node.Parent().ReplaceChild(node.Parent(), node, t)
		return
	}

}

func parseCitation(block text.Reader, pc parser.Context) (int, int) {
	block.SkipSpaces()
	line, segment := block.PeekLine()
	if line == nil {
		return -1, -1
	}
	startLine, _ := block.Position()
	width, pos := util.IndentWidth(line, 0)
	if width > 3 {
		return -1, -1
	}
	if width != 0 {
		pos++
	}
	if line[pos] != '[' {
		return -1, -1
	}
	open := segment.Start + pos + 1
	closes := -1
	block.Advance(pos + 1)
	for {
		line, segment = block.PeekLine()
		if line == nil {
			return -1, -1
		}
		closure := util.FindClosure(line, '[', ']', false, false)
		if closure > -1 {
			closes = segment.Start + closure
			next := closure + 1
			if next >= len(line) || line[next] != ':' {
				return -1, -1
			}
			block.Advance(next + 1)
			break
		}
		block.AdvanceLine()
	}
	if closes < 0 {
		return -1, -1
	}
	label := block.Value(text.NewSegment(open, closes))
	if util.IsBlank(label) {
		return -1, -1
	}
	block.SkipSpaces()
	destination, ok := parseLinkDestination(block)
	if !ok {
		return -1, -1
	}
	line, segment = block.PeekLine()
	isNewLine := line == nil || util.IsBlank(line)

	endLine, _ := block.Position()
	_, spaces, _ := block.SkipSpaces()
	opener := block.Peek()
	if opener != '"' && opener != '\'' && opener != '(' {
		if !isNewLine {
			return -1, -1
		}
		ref := parser.NewReference(label, destination, nil)
		pc.AddReference(ref)
		return startLine, endLine + 1
	}
	if spaces == 0 {
		return -1, -1
	}
	block.Advance(1)
	open = -1
	closes = -1
	closer := opener
	if opener == '(' {
		closer = ')'
	}
	for {
		line, segment = block.PeekLine()
		if line == nil {
			return -1, -1
		}
		if open < 0 {
			open = segment.Start
		}
		closure := util.FindClosure(line, opener, closer, false, true)
		if closure > -1 {
			closes = segment.Start + closure
			block.Advance(closure + 1)
			break
		}
		block.AdvanceLine()
	}
	if closes < 0 {
		return -1, -1
	}

	line, segment = block.PeekLine()
	if line != nil && !util.IsBlank(line) {
		if !isNewLine {
			return -1, -1
		}
		title := block.Value(text.NewSegment(open, closes))
		ref := parser.NewReference(label, destination, title)
		pc.AddReference(ref)
		return startLine, endLine
	}

	title := block.Value(text.NewSegment(open, closes))

	endLine, _ = block.Position()
	ref := parser.NewReference(label, destination, title)
	pc.AddReference(ref)
	return startLine, endLine + 1
}

func parseLinkDestination(block text.Reader) ([]byte, bool) {
	block.SkipSpaces()
	line, _ := block.PeekLine()
	buf := []byte{}
	if block.Peek() == '<' {
		i := 1
		for i < len(line) {
			c := line[i]
			if c == '\\' && i < len(line)-1 && util.IsPunct(line[i+1]) {
				buf = append(buf, '\\', line[i+1])
				i += 2
				continue
			} else if c == '>' {
				block.Advance(i + 1)
				return line[1:i], true
			}
			buf = append(buf, c)
			i++
		}
		return nil, false
	}
	opened := 0
	i := 0
	for i < len(line) {
		c := line[i]
		if c == '\\' && i < len(line)-1 && util.IsPunct(line[i+1]) {
			buf = append(buf, '\\', line[i+1])
			i += 2
			continue
		} else if c == '(' {
			opened++
		} else if c == ')' {
			opened--
			if opened < 0 {
				break
			}
		} else if util.IsSpace(c) {
			break
		}
		buf = append(buf, c)
		i++
	}
	block.Advance(i)
	return line[:i], len(line[:i]) != 0
}

type CitationExt struct{}

func NewCitationExt() *CitationExt {
	return &CitationExt{}
}

type citationRenderer struct{}

func (c citationRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCitation, c.render)
}

func (c citationRenderer) render(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (sc *CitationExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithParagraphTransformers(
			// Must be less than link ref, which is 100.
			util.Prioritized(citationParagraphTransformer{}, 99)))

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(citationRenderer{}, 999)))
}
