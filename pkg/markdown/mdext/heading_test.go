package mdext

import (
	"fmt"
	"github.com/jschaf/b2/pkg/htmls/tags"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
	"testing"
)

func anchoredH1(id, content string) string {
	return tags.H1Attrs("id="+id,
		content,
		tags.AAttrs(
			fmt.Sprintf(`class=heading-anchor href="#%s"`, id),
			"¶"))
}

func anchoredH2(id, content string) string {
	return tags.H2Attrs("id="+id,
		content,
		tags.AAttrs(
			fmt.Sprintf(`class=heading-anchor href="#%s"`, id),
			"¶"))
}

func TestNewHeadingExt_HeadingAnchorStyleShow(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			texts.Dedent(`
				# h1.1
				## h2.1
     `),
			tags.Join(anchoredH1("h1.1", "h1.1"), anchoredH2("h2.1", "h2.1")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(
				t, NewHeadingIDExt(), NewHeadingExt(HeadingAnchorStyleShow))
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}

func TestNewHeadingExt_HeadingAnchorStyleNone(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			texts.Dedent(`
				# h1.1
				## h2.1
     `),
			tags.Join(
				tags.H1Attrs("id=h1.1", "h1.1"),
				tags.H2Attrs("id=h2.1", "h2.1"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(
				t, NewHeadingIDExt(), NewHeadingExt(HeadingAnchorStyleNone))
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
