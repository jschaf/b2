package mdext

import (
	"testing"

	"github.com/jschaf/jsc/pkg/markdown/mdctx"
	"github.com/jschaf/jsc/pkg/markdown/mdtest"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/jsc/pkg/texts"
)

func TestArticleExt(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		want      string
		wantTitle string
	}{
		{
			name: "title with math",
			src: texts.Dedent(`
				# Models for $3^3a^k$

				foo
				`),
			want: texts.Dedent(`
					<article>
            <header>
							<time datetime="0001-01-01">January  1, 0001</time>
							<h1 class="title"><a href="" title="Models for 3³aᵏ">Models for $3^3a^k$</a></h1>
            </header>
						<p>foo</p>
					</article>`),
			wantTitle: "Models for 3³aᵏ",
		},
		{
			name: "h1 + p",
			src: texts.Dedent(`
				# header
				foo

				bar`),
			want: texts.Dedent(`
					<article>
            <header>
							<time datetime="0001-01-01">January  1, 0001</time>
							<h1 class="title"><a href="" title="header">header</a></h1>
            </header>
						<p>foo</p>
						<p>bar</p>
					</article>`),
			wantTitle: "header",
		},
		{
			name: "h1 italic + p",
			src: texts.Dedent(`
				# *header*
				foo

				bar`),
			want: texts.Dedent(`
					<article>
            <header>
							<time datetime="0001-01-01">January  1, 0001</time>
							<h1 class="title"><a href="" title="header"><em>header</em></a></h1>
            </header>
						<p>foo</p>
						<p>bar</p>
					</article>`),
			wantTitle: "header",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewArticleExt(), NewTimeExt(), NewHeaderExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
			got := mdctx.GetTitle(ctx)
			if diff := cmp.Diff(got.Text, tt.wantTitle); diff != "" {
				t.Errorf("Article title mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
