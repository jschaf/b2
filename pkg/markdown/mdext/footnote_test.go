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
        [^side:foo] link-text

        ::: footnote side:foo
        body-text
        :::
      `),
			texts.Dedent(`
        <p>
          <span class="footnote-link" id="footnote-link-side:foo" role="doc-noteref">
            <a href="#footnote-body-side:foo">[FN]</a>
          </span>
          link-text
        </p>
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote">
          <p>body-text</p>
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
          <span class="footnote-link" id="footnote-link-side:foo" role="doc-noteref">
            <a href="#footnote-body-side:foo">[FN]</a>
          </span>
          link-text
        </p>
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote">
          <p>body-text</p>
        </aside>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewFootnoteExt(), NewColonBlockExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
