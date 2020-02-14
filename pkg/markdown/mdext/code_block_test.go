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
			"h1 + p",
			texts.Dedent("``` go\n" +
				"func foo() {}\n" +
				"```\n"),
			texts.Dedent(`
					<pre>
						<code class="lang-go">
							func foo() {}
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
