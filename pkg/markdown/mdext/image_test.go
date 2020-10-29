package mdext

import (
	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"testing"

	"github.com/jschaf/b2/pkg/texts"
)

func TestNewImageExt(t *testing.T) {
	const path = "/home/joe/file.md"
	tests := []struct {
		name       string
		src        string
		want       string
		wantAssets []assets.Blob
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
			[]assets.Blob{
				{Src: "/home/joe/qux.png", Dest: "qux.png"},
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
			[]assets.Blob{
				{Src: "/home/joe/qux.png", Dest: "/some_slug/qux.png"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewTOMLExt(), NewImageExt())
			mdctx.SetFilePath(ctx, path)
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
			if diff := cmp.Diff(tt.wantAssets, mdctx.GetAssets(ctx)); diff != "" {
				t.Fatalf("assets context mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
