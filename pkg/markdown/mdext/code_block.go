package mdext

import (
	"bytes"
	"fmt"
	"html"
	"io"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type codeBlockRenderer struct{}

func (c *codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, c.render)
}

func (c *codeBlockRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	n := node.(*ast.FencedCodeBlock)

	if entering {
		language := n.Language(source)

		lexer := getLexer(language)
		formatter := NewCodeBlockFormatter()

		tokenIter, err := lexer.Tokenise(nil, c.readAllLines(n, source))
		if err != nil {
			panic(err)
		}
		if err := formatter.Format(w, tokenIter); err != nil {
			panic(err)
		}

	}
	return ast.WalkContinue, nil
}

func (c *codeBlockRenderer) readAllLines(n *ast.FencedCodeBlock, source []byte) string {
	var b bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		b.Write(line.Value(source))
	}
	return b.String()
}

func getLexer(language []byte) chroma.Lexer {
	lexer := lexers.Fallback
	if language != nil {
		lexer = lexers.Get(string(language))
	}
	lexer = chroma.Coalesce(lexer)
	return lexer
}

func NewCodeBlockRenderer() *codeBlockRenderer {
	return &codeBlockRenderer{}
}

// codeBlockExt is a Goldmark extension to register the AST transformer and
// renderer
type codeBlockExt struct{}

func NewCodeBlockExt() *codeBlockExt {
	return &codeBlockExt{}
}

func (c *codeBlockExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewCodeBlockRenderer(), 999),
		),
	)
}

type codeBlockFormatter struct {
}

func NewCodeBlockFormatter() *codeBlockFormatter {
	return &codeBlockFormatter{}
}

func (c *codeBlockFormatter) Format(w io.Writer, iterator chroma.Iterator) error {
	fmt.Fprintf(w, "<code-block-container style='display:block'>")
	fmt.Fprintf(w, "<code-block style='white-space:pre; display:block;'>")

	tokens := iterator.Tokens()
	lines := chroma.SplitTokensIntoLines(tokens)
	for _, tokens := range lines {
		for _, token := range tokens {
			h := html.EscapeString(token.String())
			switch token.Type {

			case chroma.Comment:
				fallthrough
			case chroma.CommentHashbang:
				fallthrough
			case chroma.CommentMultiline:
				fallthrough
			case chroma.CommentPreproc:
				fallthrough
			case chroma.CommentPreprocFile:
				fallthrough
			case chroma.CommentSingle:
				fallthrough
			case chroma.CommentSpecial:
				fmt.Fprintf(w, "<code-comment>%s</code-comment>", h)

			case chroma.Keyword:
				fallthrough
			case chroma.KeywordConstant:
				fallthrough
			case chroma.KeywordDeclaration:
				fallthrough
			case chroma.KeywordNamespace:
				fallthrough
			case chroma.KeywordPseudo:
				fallthrough
			case chroma.KeywordReserved:
				fallthrough
			case chroma.KeywordType:
				fmt.Fprintf(w, "<code-kw>%s</code-kw>", h)

			case chroma.String:
				fallthrough
			case chroma.StringAffix:
				fallthrough
			case chroma.StringBacktick:
				fallthrough
			case chroma.StringChar:
				fallthrough
			case chroma.StringDelimiter:
				fallthrough
			case chroma.StringDoc:
				fallthrough
			case chroma.StringDouble:
				fallthrough
			case chroma.StringEscape:
				fallthrough
			case chroma.StringHeredoc:
				fallthrough
			case chroma.StringInterpol:
				fallthrough
			case chroma.StringOther:
				fallthrough
			case chroma.StringRegex:
				fallthrough
			case chroma.StringSingle:
				fallthrough
			case chroma.StringSymbol:
				fmt.Fprintf(w, "<code-str>%s</code-str>", h)

			default:
				fmt.Fprintf(w, h)

			}
		}
	}

	fmt.Fprintf(w, "</code-block>")
	fmt.Fprintf(w, "</code-block-container>")
	return nil
}
