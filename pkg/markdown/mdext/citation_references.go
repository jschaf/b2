package mdext

import (
	"github.com/yuin/goldmark/ast"
)

var KindCitationReferences = ast.NewNodeKind("citation_references")

// CitationReferences is a list of citations that appeared in the document.
type CitationReferences struct {
	ast.BaseBlock
	// All citations from the source document. Ordered by appearance.
	Citations []*Citation
}

func NewCitationReferences() *CitationReferences {
	return &CitationReferences{}
}

func (r *CitationReferences) Kind() ast.NodeKind {
	return KindCitationReferences
}

func (r *CitationReferences) Dump(source []byte, level int) {
	ast.DumpHelper(r, source, level, nil, nil)
}
