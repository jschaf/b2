package mdext

import (
	"github.com/jschaf/b2/pkg/htmls/tags"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
	"testing"
)

func TestNewParagraphExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			texts.Dedent(`
				foo bar baz
     `),
			tags.P("foo bar baz"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewParagraphExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
