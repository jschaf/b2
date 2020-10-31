package mdext

import (
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
	"testing"
)

func TestNewFootnoteExt(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"immediately after",
			texts.Dedent(`
        [^side:foo] 
        alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima

        ::: footnote side:foo
        body-text
        :::
      `),
			texts.Dedent(`
        <p>
          <a class="footnote-link" role="doc-noteref" href="#footnote-body-side:foo" id="footnote-link-side:foo">
			      <cite>[1]</cite>
			    </a>
          alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima
        </p>
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote" style="margin-top: -54px">
          <p><cite>[1]</cite> body-text</p>
        </aside>
      `),
		},
		{
			"immediately before",
			texts.Dedent(`
        ::: footnote side:foo
        body-text
        :::

        [^side:foo] link-text
      `),
			texts.Dedent(`
        <p>
          <a class="footnote-link" role="doc-noteref" href="#footnote-body-side:foo" id="footnote-link-side:foo">
			      <cite>[1]</cite>
			    </a>
          link-text
        </p>
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote" style="margin-top: -18px">
          <p><cite>[1]</cite> body-text</p>
        </aside>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewFootnoteExt(),
				NewColonBlockExt(),
				NewCustomExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
