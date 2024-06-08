package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
)

func TestNewHeadingIDExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			texts.Dedent(`
				# h1
				# h1 dupe
				# h1 dupe
				## h2 dupe
				## h2
				## h2 dupe
			`),
			texts.Dedent(`
				<h1 id="h1">h1</h1>
				<h1 id="h1-dupe">h1 dupe</h1>
				<h1 id="h1-dupe-1">h1 dupe</h1>
				<h2 id="h2-dupe">h2 dupe</h2>
				<h2 id="h2">h2</h2>
				<h2 id="h2-dupe-1">h2 dupe</h2>
		`),
		},
		{
			`# h1--   joe`,
			`<h1 id="h1-joe">h1--   joe</h1>`,
		},
		{
			`## Inverted indexes for experiment IDs`,
			`<h2 id="inverted-indexes-for-experiment-ids">Inverted indexes for experiment IDs</h2>`,
		},
		{
			`## 2. strip leading nums`,
			`<h2 id="strip-leading-nums">2. strip leading nums</h2>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewHeadingIDExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
