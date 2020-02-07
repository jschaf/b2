package mdext

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type articleHeaderASTTrans struct {
}

func (a articleHeaderASTTrans) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	heading := firstHeading(doc, reader.Source())
	if heading == nil {
		return
	}

	parent := heading.Parent()
	if parent == nil {
		return
	}
	meta := GetTOMLMeta(pc)

	link := ast.NewLink()
	link.Title = []byte(meta.Title)
	link.Destination = []byte("/fffbar" + meta.Slug)
	link.AppendChild(link, heading)

	parent.InsertAfter(parent, heading, link)
}

func firstHeading(doc *ast.Document, source []byte) ast.Node {
	var hNode ast.Node
	//var txt string
	err := ast.Walk(doc, func(n ast.Node, entering bool) (status ast.WalkStatus, err error) {
		bytes := string(n.Text(source))
		fmt.Println(bytes)
		if n.Kind() == ast.KindHeading {
			hNode = n
			//txt = bytes
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil
	}
	return hNode
}

var defaultArticleHeaderASTTrans = &articleHeaderASTTrans{}

func NewArticleHeadingASTTransformer() parser.ASTTransformer {
	return defaultArticleHeaderASTTrans
}

type articleHeading struct {
}

func (a *articleHeading) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(NewArticleHeadingASTTransformer(), 999),
		),
	)
}

var ArticleHeading = &articleHeading{}
