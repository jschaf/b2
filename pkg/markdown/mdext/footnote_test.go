package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
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
          <a href="#footnote-body-side:foo" class="footnote-link" role="doc-noteref" id="footnote-link-side:foo">
			      <cite>[1]</cite>
			    </a>
          alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima
        </p>
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote" style="margin-top: -54px">
          <p><cite class=cite-inline>[1]</cite> body-text</p>
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
          <a href="#footnote-body-side:foo" class="footnote-link" role="doc-noteref" id="footnote-link-side:foo">
			      <cite>[1]</cite>
			    </a>
          link-text
        </p>
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote" style="margin-top: -18px">
          <p><cite class=cite-inline>[1]</cite> body-text</p>
        </aside>
      `),
		},
		{
			"margin note",
			texts.Dedent(`
        [^margin:foo] alpha bravo charlie

        ::: footnote margin:foo
        body-text
        :::
      `),
			texts.Dedent(`
        <p>
          <a href="#footnote-body-margin:foo" class="footnote-link" role="doc-noteref" id="footnote-link-margin:foo"></a>
          alpha bravo charlie
        </p>
        <aside class="footnote-body" id="footnote-body-margin:foo" role="doc-endnote" style="margin-top: -18px">
          <p>body-text</p>
        </aside>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewFootnoteExt(cite.IEEE, NewCitationNopAttacher()),
				NewColonBlockExt(),
				NewCustomExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
