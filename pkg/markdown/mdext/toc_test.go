package mdext

import (
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
						<li>h2.1</li>
						<li>
							<ol class="toc-list toc-level-3">
								<li>h3.1</li>
							</ol>
						</li>
						<li>h2.2</li>
					</ol>
				</div>
				<h1>h1.1</h1>
				<h2>h2.1</h2>
				<h3>h3.1</h3>
				<h2>h2.2</h2>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := newMdTester(t, NewColonLineExt(), NewTOCExt(TOCStyleShow))
			doc := mustParseMarkdown(t, md, ctx, tt.src)
			assertNoRenderDiff(t, doc, md, tt.src, tt.want)
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
			md, ctx := newMdTester(t, NewColonLineExt(), NewTOCExt(TOCStyleNone))
			doc := mustParseMarkdown(t, md, ctx, tt.src)
			assertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
