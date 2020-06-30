package mdext

import (
	"fmt"
	"github.com/jschaf/b2/pkg/htmls/tags"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
	"testing"
)

func tocList(level int, ts ...string) string {
	return tags.OlAttrs(
		fmt.Sprintf(`class="toc-list toc-level-%d"`, level),
		ts...)
}

func tocItem(count, id, title string) string {
	return tags.Join(
		tags.SpanAttrs("class=toc-ordering", count),
		tags.AAttrs(`href="#`+id+`"`, title))
}

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
			tags.Join(
				tags.DivAttrs("class=toc",
					tocList(2,
						tocItem("1", "h2.1", "h2.1"),
						tocList(3, tocItem("1.1", "h3.1", "h3.1")),
						tocItem("2", "h2.2", "h2.2"),
					),
				),
				tags.H1Attrs("id=h1.1", "h1.1"),
				tags.H2Attrs("id=h2.1", "h2.1"),
				tags.H3Attrs("id=h3.1", "h3.1"),
				tags.H2Attrs("id=h2.2", "h2.2"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewColonLineExt(), NewTOCExt(TOCStyleShow), NewHeadingIDExt(),
				NewColonBlockExt())
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
