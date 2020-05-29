package mdext

import (
	"github.com/jschaf/b2/pkg/cite/bibtex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var KindReferenceList = ast.NewNodeKind("reference_list")
var KindReference = ast.NewNodeKind("reference")

// ReferenceList is a list of references cited in a document.
type ReferenceList struct {
	ast.BaseBlock
}

func NewReferenceList() *ReferenceList {
	return &ReferenceList{}
}

func (r *ReferenceList) Kind() ast.NodeKind {
	return KindReferenceList
}

func (r *ReferenceList) Dump(source []byte, level int) {
	ast.DumpHelper(r, source, level, nil, nil)
}

// Reference is a single item in a reference list.
type Reference struct {
	ast.BaseBlock
	CiteID bibtex.Key
}

func NewReference() *Reference {
	return &Reference{}
}

func (r *Reference) Kind() ast.NodeKind {
	return KindReference
}

func (r *Reference) Dump(source []byte, level int) {
	ast.DumpHelper(r, source, level, nil, nil)
}

// referenceRenderer renders Reference and ReferenceList nodes.
type referenceRenderer struct {
}

func (r *referenceRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindReference, r.renderReference)
	reg.Register(KindReferenceList, r.renderReferenceList)
}

func (r *referenceRenderer) renderReference(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	ref := n.(*Reference)
	if entering {
		_, _ = w.WriteString(`<div class="reference">`)
		_, _ = w.WriteString(ref.CiteID)
	} else {
		_, _ = w.WriteString(`</div>`)
	}
	return ast.WalkContinue, nil
}

func (r *referenceRenderer) renderReferenceList(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<div class="reference-list">`)
	} else {
		_, _ = w.WriteString(`</div>`)
	}
	return ast.WalkContinue, nil
}

type ReferenceListExt struct {
}

func NewReferenceListExt() *ReferenceListExt {
	return &ReferenceListExt{}
}

func (r *ReferenceListExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&referenceRenderer{}, 999)))
}
