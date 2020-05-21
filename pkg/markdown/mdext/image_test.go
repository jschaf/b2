package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark/parser"
)

func TestNewImageExt(t *testing.T) {
	const path = "/home/joe/file.md"
	tests := []struct {
		name    string
		src     string
		want    string
		wantCtx map[parser.ContextKey]interface{}
	}{
		{
			"single image in a paragraph",
			texts.Dedent(`
        In a paragraph. ![alt text](./qux.png "title")`),
			texts.Dedent(`
        <p>
          In a paragraph.
          <img src="qux.png" alt="alt text" title="title">
        </p>
     `),
			map[parser.ContextKey]interface{}{
				assetsCtxKey: map[string]string{"qux.png": "/home/joe/qux.png"},
			},
		},
		{
			"single image in a paragraph with slug",
			texts.Dedent(`
				+++
				slug = "some_slug"
				+++

				In a paragraph. ![alt text](./qux.png "title")
      `),
			texts.Dedent(`
        <p>
          In a paragraph.
          <img src="/some_slug/qux.png" alt="alt text" title="title">
        </p>
     `),
			map[parser.ContextKey]interface{}{
				assetsCtxKey: map[string]string{"/some_slug/qux.png": "/home/joe/qux.png"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := newMdTester(t, NewTOMLExt(), NewImageExt())
			SetFilePath(ctx, path)
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)

			assertCtxContainsAll(t, ctx, tt.wantCtx)
		})
	}
}
