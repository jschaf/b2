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
					<div class="code-block-container">
						<pre class="code-block">
							<code-kw>func</code-kw> <code-fn>foo</code-fn>() {}
						</pre>
					</div>
    `),
		},
		{
			"go func",
			texts.Dedent("``` go\n" +
				"Foo 28%\n" +
				"```\n"),
			texts.Dedent(`
					<div class="code-block-container">
						<pre class="code-block">
							Foo 28%
						</pre>
					</div>
     `),
		},
		{
			"go func receiver",
			texts.Dedent("``` go\n" +
				"func (t *T) foo() {}\n" +
				"```\n"),
			texts.Dedent(`
					<div class="code-block-container">
						<pre class="code-block">
							<code-kw>func</code-kw> (t *T) <code-fn>foo</code-fn>() {}
						</pre>
					</div>
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
