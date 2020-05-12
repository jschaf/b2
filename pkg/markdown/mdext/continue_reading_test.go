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
			"uses CONTINUE_READING token",
			"my-slug",
			texts.Dedent(`
       # title
		
       foo bar
		
       qux
		
       CONTINUE_READING
		
       baz
     `),
			texts.Dedent(`
				 <h1>title</h1>
         <p>foo bar</p>
         <p>qux</p>
    ` + contReadingLink("/my-slug")),
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
				t.Errorf("ContinueReading mismatch (-want +got)\n%s", diff)
			}
		})
	}
}

func TestNopContinueReadingTransformer(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"removes CONTINUE_READING token",
			texts.Dedent(`
       # title
		
       foo bar
		
       qux
		
       CONTINUE_READING
		
       baz
     `),
			texts.Dedent(`
				 <h1>title</h1>
         <p>foo bar</p>
         <p>qux</p>
         <p>baz</p>
    `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewNopContinueReadingExt(),
			))
			buf := new(bytes.Buffer)
			ctx := parser.NewContext()
			if err := md.Convert([]byte(tt.src), buf, parser.WithContext(ctx)); err != nil {
				t.Fatal(err)
			}

			if diff, err := htmls.Diff(buf, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			} else if diff != "" {
				t.Errorf("ContinueReading mismatch (-want +got)\n%s", diff)
			}
		})
	}
}
