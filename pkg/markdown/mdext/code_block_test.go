package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/markdown/mdtest"

	"github.com/jschaf/b2/pkg/texts"
)

func TestCodeBlockExt(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"go func",
			texts.Dedent("``` go\n" +
				"func foo() {}\n" +
				"```\n"),
			texts.Dedent(`
					<fieldset class="code-block-container">
						<pre class="code-block">
							<code-kw>func</code-kw> <code-fn>foo</code-fn>() {}
						</pre>
					</fieldset>
    `),
		},
		{
			"go func",
			texts.Dedent("``` go\n" +
				"Foo 28%\n" +
				"```\n"),
			texts.Dedent(`
					<fieldset class="code-block-container">
						<pre class="code-block">
							Foo 28%
						</pre>
					</fieldset>
     `),
		},
		{
			"go func receiver",
			texts.Dedent("``` go\n" +
				"func (t *T) foo() {}\n" +
				"```\n"),
			texts.Dedent(`
					<fieldset class="code-block-container">
						<pre class="code-block">
							<code-kw>func</code-kw> (t *T) <code-fn>foo</code-fn>() {}
						</pre>
					</fieldset>
     `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewCodeBlockExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
