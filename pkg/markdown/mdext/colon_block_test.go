package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/markdown/mdtest"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/texts"
)

func TestNewColonBlockExt_preview(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		want       string
		wantCtxURL string
	}{
		{
			"h1 + p + colon",
			texts.Dedent(`
        # header
        foo

        ::: preview   http://example.com  
        qux
        :::
      `),
			texts.Dedent(`
        <h1>header</h1>
			  <p>foo</p>
      `),
			"http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewColonBlockExt(), NewColonLineExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
			gotPreview, _ := GetPreview(ctx, tt.wantCtxURL)
			if diff := cmp.Diff(tt.wantCtxURL, gotPreview.URL); diff != "" {
				t.Errorf("Colon block preview mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
