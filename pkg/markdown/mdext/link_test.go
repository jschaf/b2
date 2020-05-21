package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark/parser"
)

func TestNewLinkExt(t *testing.T) {
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
				assetsCtxKey: map[string]string{"paper.pdf": "/home/joe/paper.pdf"},
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
				assetsCtxKey: map[string]string{"/some_slug/paper.pdf": "/home/joe/paper.pdf"},
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
		{
			"link with preview",
			texts.Dedent(`
				[wiki link](https://en.wikipedia.org/wiki/Wiki)

				::: preview https://en.wikipedia.org/wiki/Wiki
				preview title

				foo bar
				:::
      `),
			texts.Dedent(`
        <p>
          <a href="https://en.wikipedia.org/wiki/Wiki" 
             data-link-type=wikipedia
             class="preview-target"
             data-preview-title="<p class=&quot;preview-title&quot;><a href=&quot;https://en.wikipedia.org/wiki/Wiki&quot;>preview title</a></p>" 
             data-preview-snippet="<p>foo bar</p>">
          wiki link
          </a>
        </p>
     `),
			map[parser.ContextKey]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := newMdTester(t, NewColonBlockExt(), NewTOMLExt(), NewLinkExt())
			SetFilePath(ctx, path)

			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
			assertCtxContainsAll(t, ctx, tt.wantCtx)
		})
	}
}
