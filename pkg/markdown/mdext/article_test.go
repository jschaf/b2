package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/mdtest"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/texts"
)

func TestArticleExt(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		want      string
		wantTitle string
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
			"header",
		},
		{
			"h1 italic + p",
			texts.Dedent(`
				# *header*
				foo

				bar`),
			texts.Dedent(`
					<article>
            <header>
							<time datetime="0001-01-01">January  1, 0001</time>
							<h1 class="title"><a href="" title="header"><em>header</em></a></h1>
            </header>
						<p>foo</p>
						<p>bar</p>
					</article>`),
			"header",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewArticleExt(),
				NewTimeExt(),
				NewHeaderExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
			got := mdctx.GetTitle(ctx)
			if diff := cmp.Diff(got.Text, tt.wantTitle); diff != "" {
				t.Errorf("Article title mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
