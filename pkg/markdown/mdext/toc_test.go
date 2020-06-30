package mdext

import (
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
	"testing"
)

func TestNewTOCExt_TOCStyleShow(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			texts.Dedent(`
				:toc:
			
				# h1.1
				## h2.1
				### h3.1
				## h2.2
     `),
			texts.Dedent(`
				<div class="toc">
					<ol class="toc-list toc-level-2">
						<li><span class=toc-ordering>1</span> h2.1</li>
						<li>
							<ol class="toc-list toc-level-3">
								<li><span class=toc-ordering>1.1</span> h3.1</li>
							</ol>
						</li>
						<li><span class=toc-ordering>2</span> h2.2</li>
					</ol>
				</div>
				<h1 id=h1-1>h1.1</h1>
				<h2 id=h2-1>h2.1</h2>
				<h3 id=h3-1>h3.1</h3>
				<h2 id=h2-2>h2.2</h2>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewColonLineExt(), NewTOCExt(TOCStyleShow), NewHeadingIDExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}

func TestNewTOCExt_TOCStyleNone(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			texts.Dedent(`
				:toc:
			
				# h1.1
				## h2.1
				### h3.1
				## h2.2
     `),
			texts.Dedent(`
				<h1>h1.1</h1>
				<h2>h2.1</h2>
				<h3>h3.1</h3>
				<h2>h2.2</h2>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewColonLineExt(), NewTOCExt(TOCStyleNone))
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
