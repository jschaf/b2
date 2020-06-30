package mdext

import (
	"bytes"
	"html"
	"io"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// codeBlockRenderer renders code blocks, replacing the default renderer.
type codeBlockRenderer struct{}

func (c codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, c.render)
}

func (c codeBlockRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	n := node.(*ast.FencedCodeBlock)

	if entering {
		lang := string(n.Language(source))

		lexer := getLexer(lang)

		tokenIter, err := lexer.Tokenise(nil, readAllCodeBlockLines(n, source))
		if err != nil {
			panic(err)
		}
		if err := formatCodeBlock(w, tokenIter, lang); err != nil {
			panic(err)
		}

	}
	return ast.WalkContinue, nil
}

func readAllCodeBlockLines(n *ast.FencedCodeBlock, source []byte) string {
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

func formatCodeBlock(w io.Writer, iterator chroma.Iterator, lang string) error {
	writeStrings(w, "<div class='code-block-container'>")
	writeStrings(w, "<pre class='code-block'>")

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
				writeStrings(w, "<code-comment>", h, "</code-comment>")

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
				writeStrings(w, "<code-kw>", h, "</code-kw>")

			case chroma.NameFunction:
				switch lang {
				case "go":
					if i < 2 {
						writeStrings(w, h)
						continue
					}
					isFunc := tokens[i-2].Value == "func"
					isReceiver := tokens[i-2].Value == ")"
					if isFunc || isReceiver {
						writeStrings(w, "<code-fn>", h, "</code-fn>")
					} else {
						writeStrings(w, h)
					}

				default:
					writeStrings(w, "<code-fn>", h, "</code-fn>")
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
				writeStrings(w, "<code-str>", h, "</code-str>")

			default:
				writeStrings(w, h)
			}
		}
	}

	writeStrings(w, "</pre>")
	writeStrings(w, "</div>")
	return nil
}

func writeStrings(w io.Writer, ss ...string) {
	for _, s := range ss {
		_, _ = w.Write([]byte(s))
	}
}

// CodeBlockExt extends markdown to better render code blocks with syntax
// highlighting.
type CodeBlockExt struct{}

func NewCodeBlockExt() CodeBlockExt {
	return CodeBlockExt{}
}

func (c CodeBlockExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(codeBlockRenderer{}, 999),
		),
	)
}
