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

func TestArticleExt(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"h1 + p",
			texts.Dedent(`
				# header
				foo

				bar`),
			texts.Dedent(`
					<article>
            <header>
							<time datetime="0001-01-01">January  1, 0001</time>
							<h1 class="title"><a href="" title="header">header</a></h1>
            </header>
						<p>foo</p>
						<p>bar</p>
					</article>`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewArticleExt(),
				NewTimeExt(),
				NewHeaderExt(),
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
