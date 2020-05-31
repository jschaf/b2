package mdext

import (
	"fmt"

	"github.com/jschaf/b2/pkg/cite/bibtex"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// citationRendererIEEE renders an IEEE citation.
type citationRendererIEEE struct {
	nextNum  int
	citeNums map[bibtex.Key]int
}

func (cr *citationRendererIEEE) renderCitation(writer util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	c := n.(*Citation)
	// For IEEE style we need to dedupe the citation order. The raw order
	// assigns multiple numbers for the same cite key.
	num, ok := cr.citeNums[c.Key]
	if !ok {
		num = cr.nextNum
		cr.citeNums[c.Key] = num
		cr.nextNum += 1
	}

	_, _ = writer.WriteString(
		fmt.Sprintf(`<cite id=%s data-cite-key="%s">[%d]</cite>`, c.ID(), c.Key, num))
	// Citations should generate content solely from the citation, not children.
	return ast.WalkSkipChildren, nil
}

func (cr *citationRendererIEEE) renderReferenceList(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	panic("implement me")
}
