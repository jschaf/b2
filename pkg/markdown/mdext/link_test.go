package mdext

import (
	"github.com/jschaf/b2/pkg/htmls/tags"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"testing"

	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark/parser"
)

func TestNewLinkExt_context(t *testing.T) {
	const path = "/home/joe/file.md"
	tests := []struct {
		name    string
		src     string
		want    string
		wantCtx map[parser.ContextKey]interface{}
	}{
		{
			"single relative link",
			texts.Dedent(`
				Paper: [Gorilla Title][gorilla]
		
				[gorilla]: paper.pdf
    `),
			texts.Dedent(`
      <p>
        Paper: <a href="paper.pdf" data-link-type=pdf>Gorilla Title</a>
      </p>
    `),
			map[parser.ContextKey]interface{}{
				mdctx.AssetsCtxKey: map[string]string{"paper.pdf": "/home/joe/paper.pdf"},
			},
		},
		{
			"single relative link with slug",
			texts.Dedent(`
				+++
				slug = "some_slug"
				+++
		
				Paper: [Gorilla Title][gorilla]
		
				[gorilla]: paper.pdf
    `),
			texts.Dedent(`
      <p>
        Paper: <a href="/some_slug/paper.pdf" data-link-type=pdf>Gorilla Title</a>
      </p>
    `),
			map[parser.ContextKey]interface{}{
				mdctx.AssetsCtxKey: map[string]string{"/some_slug/paper.pdf": "/home/joe/paper.pdf"},
			},
		},
		{
			"single absolute link with slug",
			texts.Dedent(`
				+++
				slug = "some_slug"
				+++
		
				Paper: [Gorilla Title][gorilla]
		
				[gorilla]: http://example.com/paper.pdf
     `),
			texts.Dedent(`
       <p>
         Paper: <a href="http://example.com/paper.pdf" data-link-type=pdf>Gorilla Title</a>
       </p>
    `),
			map[parser.ContextKey]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewColonBlockExt(), NewTOMLExt(), NewLinkExt(), NewParagraphExt())
			mdctx.SetFilePath(ctx, path)

			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
			mdtest.AssertCtxContainsAll(t, ctx, tt.wantCtx)
		})
	}
}

func TestNewLinkExt_Preview(t *testing.T) {
	const path = "/home/joe/file.md"
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"link with preview",
			texts.Dedent(`
				[wiki link](https://en.wikipedia.org/wiki/Wiki)

				::: preview https://en.wikipedia.org/wiki/Wiki
				preview title

				foo bar
				:::
      `),
			tags.Join(
				tags.P(
					tags.AAttrs(
						tags.Attrs(
							`href="https://en.wikipedia.org/wiki/Wiki"`,
							"data-link-type=wikipedia",
							`class="preview-target"`,
							`data-preview-title="<div class=&quot;preview-title&quot;><a href=&quot;https://en.wikipedia.org/wiki/Wiki&quot;>preview title</a></div>"`,
							`data-preview-snippet="<p>foo bar</p>"`),
						"wiki link"),
				),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewColonBlockExt(), NewTOMLExt(), NewLinkExt(), NewParagraphExt())
			mdctx.SetFilePath(ctx, path)

			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
