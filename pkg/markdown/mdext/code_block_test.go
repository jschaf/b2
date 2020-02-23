package mdext

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/htmls"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

func TestCodeBlockExt(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"go func",
			texts.Dedent("``` go\n" +
				"func foo() {}\n" +
				"```\n"),
			texts.Dedent(`
					<code-block-container style="display:block">
						<code-block style="white-space:pre; display:block;">
							<code-kw>func</code-kw> <code-fn>foo</code-fn>() {}
						</code
					</pre>
     `),
		},
		{
			"go func receiver",
			texts.Dedent("``` go\n" +
				"func (t *T) foo() {}\n" +
				"```\n"),
			texts.Dedent(`
					<code-block-container style="display:block">
						<code-block style="white-space:pre; display:block;">
							<code-kw>func</code-kw> (t *T) <code-fn>foo</code-fn>() {}
						</code
					</pre>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewCodeBlockExt(),
			))
			buf := new(bytes.Buffer)
			ctx := parser.NewContext()
			if err := md.Convert([]byte(tt.src), buf, parser.WithContext(ctx)); err != nil {
				t.Fatal(err)
			}

			if diff, err := htmls.Diff(buf, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			} else if diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
