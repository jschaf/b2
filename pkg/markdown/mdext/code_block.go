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
		lang := string(n.Language(source))

		lexer := getLexer(lang)
		formatter := NewCodeBlockFormatter()

		tokenIter, err := lexer.Tokenise(nil, c.readAllLines(n, source))
		if err != nil {
			panic(err)
		}
		if err := formatter.Format(w, tokenIter, lang); err != nil {
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

func getLexer(language string) chroma.Lexer {
	lexer := lexers.Fallback
	if language != "" {
		lexer = lexers.Get(language)
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

func (c *codeBlockFormatter) Format(w io.Writer, iterator chroma.Iterator, lang string) error {
	fmt.Fprintf(w, "<div class='code-block-container'>")
	fmt.Fprintf(w, "<pre class='code-block'>")

	tokens := iterator.Tokens()
	lines := chroma.SplitTokensIntoLines(tokens)
	for _, tokens := range lines {
		for i, token := range tokens {
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

			case chroma.NameFunction:
				switch lang {
				case "go":
					if i < 2 {
						fmt.Fprint(w, h)
						continue
					}
					isFunc := tokens[i-2].Value == "func"
					isReceiver := tokens[i-2].Value == ")"
					if isFunc || isReceiver {
						fmt.Fprintf(w, "<code-fn>%s</code-fn>", h)
					} else {
						fmt.Fprint(w, h)
					}

				default:
					fmt.Fprintf(w, "<code-fn>%s</code-fn>", h)
				}

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

	fmt.Fprintf(w, "</pre>")
	fmt.Fprintf(w, "</div>")
	return nil
}
