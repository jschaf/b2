package mdext

import (
	"github.com/jschaf/b2/pkg/markdown/asts"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
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

// articleTransformer wraps the markdown document in an HTML article tag.
type articleTransformer struct{}

func newArticleTransformer() *articleTransformer {
	return &articleTransformer{}
}

func (at *articleTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	meta := GetTOMLMeta(pc)
	heading := firstHeading(doc)
	if heading == nil {
		panic("nil heading, file path: " + mdctx.GetFilePath(pc))
	}
	title := string(heading.Text(reader.Source()))
	mdctx.SetTitle(pc, title)

	parent := heading.Parent()
	if parent == nil {
		return
	}

	article := NewArticle()
	header := NewHeader()
	link := ast.NewLink()
	link.Title = []byte(title)
	link.Destination = []byte(meta.Path)
	asts.Reparent(link, heading)
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
	// from the parent and connects its prev sibling to the next sibling. Since we
	// use heading for location info, move it last so we don't disconnect it.
	parent.ReplaceChild(parent, heading, article)
}

func firstHeading(doc *ast.Document) ast.Node {
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
	return hNode
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
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(newArticleTransformer(), 900),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(articleRenderer{}, 999),
		),
	)
}
