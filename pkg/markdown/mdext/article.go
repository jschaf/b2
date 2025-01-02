package mdext

import (
	"errors"
	"github.com/jschaf/jsc/pkg/markdown/asts"
	"github.com/jschaf/jsc/pkg/markdown/extenders"
	"github.com/jschaf/jsc/pkg/markdown/mdctx"
	"github.com/jschaf/jsc/pkg/markdown/ord"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindArticle = ast.NewNodeKind("Article")

// Article is a block node representing an article tag in HTML.
type Article struct {
	ast.BaseBlock
}

func NewArticle() *Article {
	return &Article{}
}

func (a *Article) Dump(source []byte, level int) {
	ast.DumpHelper(a, source, level, nil, nil)
}

func (a *Article) Kind() ast.NodeKind {
	return KindArticle
}

// articleTransformer wraps the Markdown document in an HTML article tag.
type articleTransformer struct{}

func newArticleTransformer() *articleTransformer {
	return &articleTransformer{}
}

func (at *articleTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	meta := GetTOMLMeta(pc)
	heading := firstHeading(doc)
	if heading == nil {
		mdctx.PushError(pc, errors.New("no main heading in file: "+mdctx.GetFilePath(pc)))
		return
	}
	titleText := renderTextTitle(reader, heading)

	parent := heading.Parent()
	if parent == nil {
		return
	}

	article := NewArticle()
	header := NewHeader()
	link := ast.NewLink()
	link.Title = []byte(titleText)
	link.Destination = []byte(meta.Path)
	asts.Reparent(link, heading)
	mdctx.SetTitle(pc, mdctx.Title{
		Text: titleText,
		Node: link,
	})
	newHeading := ast.NewHeading(1)
	newHeading.SetAttribute([]byte("class"), []byte("title"))
	newHeading.AppendChild(newHeading, link)
	header.AppendChild(header, NewTime(meta.Date))
	header.AppendChild(header, newHeading)
	article.AppendChild(article, header)

	cur := heading.NextSibling()
	for cur != nil {
		next := cur.NextSibling()
		article.AppendChild(article, cur)
		cur = next
	}
	// This step must come last. When we move a node in Goldmark, it detaches
	// from the parent and connects its prev sibling to the next sibling. Since
	// we use heading for location info, move it last so we don't disconnect it.
	parent.ReplaceChild(parent, heading, article)
}

func firstHeading(doc *ast.Document) *ast.Heading {
	var hNode ast.Node

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() == ast.KindHeading {
			hNode = n
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil
	}
	return hNode.(*ast.Heading)
}

// articleRenderer is the HTML renderer for an article node.
type articleRenderer struct{}

func (a articleRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindArticle, a.render)
}

func (a articleRenderer) render(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if entering {
		_, _ = w.WriteString("<article>\n")
	} else {
		_, _ = w.WriteString("\n</article>\n")
	}
	return ast.WalkContinue, nil
}

// ArticleExt is a Goldmark extension to run the AST transformer and renderer.
type ArticleExt struct{}

func NewArticleExt() *ArticleExt {
	return &ArticleExt{}
}

func (a *ArticleExt) Extend(m goldmark.Markdown) {
	extenders.AddASTTransform(m, newArticleTransformer(), ord.ArticleTransformer)
	extenders.AddRenderer(m, articleRenderer{}, ord.ArticleRenderer)
}
