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

var KindArticle = ast.NewNodeKind("Article")

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

var titleCtxKey = parser.NewContextKey()

func GetTitle(pc parser.Context) string {
	return pc.Get(titleCtxKey).(string)
}

// ArticleTransformer wraps the markdown document in an HTML article tag.
type ArticleTransformer struct{}

func NewArticleTransformer() *ArticleTransformer {
	return &ArticleTransformer{}
}

func (at *ArticleTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	heading := firstHeading(doc, reader.Source())
	if heading == nil {
		panic("nil heading")
	}
	title := string(heading.Text(reader.Source()))
	fmt.Printf("title: %s\n", title)
	pc.Set(titleCtxKey, title)

	parent := heading.Parent()
	if parent == nil {
		return
	}
	meta := GetTOMLMeta(pc)

	article := NewArticle()
	link := ast.NewLink()
	link.Title = []byte(title)
	link.Destination = []byte("/" + meta.Slug)
	article.AppendChild(article, NewTime(meta.Date))
	article.AppendChild(article, link)
	for next := heading.NextSibling(); next != nil; next = next.NextSibling() {
		fmt.Println("  appending sibling: " + next.Kind().String())
		article.AppendChild(article, next)
	}
	// These step must come last. When we move a node in Goldmark, it detaches
	// from the parent and connects its prev sibling to the next sibling. Since we
	// use heading for location info, move it last so we don't disconnect it.
	parent.ReplaceChild(parent, heading, article)
	link.AppendChild(link, heading)
}

func firstHeading(doc *ast.Document, source []byte) ast.Node {
	var hNode ast.Node

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
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

// articleRenderer is the HTML renderer for an Article node.
type articleRenderer struct{}

func NewArticleRenderer() *articleRenderer {
	return &articleRenderer{}
}

func (a articleRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindArticle, a.render)
}

func (a articleRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if entering {
		_, _ = w.WriteString("<article>\n")
	} else {
		_, _ = w.WriteString("\n</article>\n")
	}
	return ast.WalkContinue, nil
}

// articleExt is a Goldmark extension to render the AST transformer and
//renderer.
type articleExt struct{}

func NewArticleExt() *articleExt {
	return &articleExt{}
}

func (a *articleExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(NewArticleTransformer(), 999),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewArticleRenderer(), 999),
		),
	)
}
