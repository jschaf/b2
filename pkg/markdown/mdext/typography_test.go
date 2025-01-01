package mdext

import (
	"testing"

	"github.com/jschaf/jsc/pkg/markdown/mdtest"

	"github.com/jschaf/jsc/pkg/htmls/tags"
)

func TestNewTypographyExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"foo--bar", tags.P("foo", enDash, "bar")},
		{"foo -- bar", tags.P("foo ", enDash, " bar")},
		{"`a--`", tags.P(tags.Code("a--"))},
		{"foo---bar", tags.P("foo", emDash, "bar")},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewTypographyExt())

			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
