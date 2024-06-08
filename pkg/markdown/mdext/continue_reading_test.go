package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/markdown/mdtest"

	"github.com/jschaf/b2/pkg/texts"
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
			md, ctx := mdtest.NewTester(t, NewContinueReadingExt())
			SetTOMLMeta(ctx, PostMeta{
				Slug: tt.slug,
			})
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
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
			md, ctx := mdtest.NewTester(t, NewNopContinueReadingExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
