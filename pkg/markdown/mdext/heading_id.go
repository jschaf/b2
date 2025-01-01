package mdext

import (
	"github.com/jschaf/jsc/pkg/markdown/asts"
	"github.com/jschaf/jsc/pkg/markdown/extenders"
	"github.com/jschaf/jsc/pkg/markdown/mdctx"
	"github.com/jschaf/jsc/pkg/markdown/ord"
	"github.com/jschaf/jsc/pkg/texts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// maxHeadingIDLen controls the length of heading IDs. 36 is a good balance
// between brevity and detail. The following phrase is 35 characters:
//
//	inverted-indexes-for-experiment-ids
const maxHeadingIDLen = 36

// headingIDTransformer is an AST transformer that adds an ID attribute to each
// heading.
type headingIDTransformer struct{}

func (h headingIDTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ids := mdctx.HeadingIDs(pc)
	_ = asts.WalkHeadings(node, func(h *ast.Heading) (ast.WalkStatus, error) {
		id := generateHeadingID(ids, h, reader.Source())
		h.SetAttribute([]byte("id"), id)
		return ast.WalkSkipChildren, nil
	})
}

func generateHeadingID(ids map[string]struct{}, h *ast.Heading, src []byte) []byte {
	b := asts.WriteSlugText(make([]byte, maxHeadingIDLen, maxHeadingIDLen+2), h, src)
	if !hasHeadingID(ids, b) {
		ids[string(b)] = struct{}{}
		return b
	}

	// Starting appending -1, -2, ... etc.
	b = append(b, '-')
	b = append(b, '0')
	last := len(b) - 1
	for i := 1; i < 10; i++ {
		d := byte('0' + i)
		b[last] = d
		if !hasHeadingID(ids, b) {
			ids[string(b)] = struct{}{}
			return b
		}
	}
	panic("no unique heading found after 10 iterations: " + string(b))
}

func hasHeadingID(ids map[string]struct{}, b []byte) bool {
	ss := texts.ReadonlyString(b)
	_, ok := ids[ss]
	return ok
}

type HeadingIDExt struct{}

func NewHeadingIDExt() *HeadingIDExt {
	return &HeadingIDExt{}
}

func (h *HeadingIDExt) Extend(m goldmark.Markdown) {
	extenders.AddASTTransform(m, headingIDTransformer{}, ord.HeadingIdTransformer)
}
