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

func TestContinueReadingTransformer(t *testing.T) {
	tests := []struct {
		name string
		slug string
		src  string
		want string
	}{
		{
			"uses CONTINUE READING token",
			"my-slug",
			texts.Dedent(`
       # title
		
       foo bar
		
       qux
		
       CONTINUE READING
		
       baz
     `),
			texts.Dedent(`
				 <h1>title</h1>
         <p>foo bar</p>
         <p>qux</p>
    ` + contReadingLink("/my-slug")),
		},
		{
			"uses first para",
			"another-slug",
			texts.Dedent(`
					# title

					foo bar

					baz
      `),
			texts.Dedent(`
					<h1>title</h1>
          <p>foo bar</p>
     ` + contReadingLink("/another-slug")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewContinueReadingExt(),
			))
			buf := new(bytes.Buffer)
			ctx := parser.NewContext()
			setTOMLMeta(ctx, PostMeta{
				Slug: tt.slug,
			})
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
