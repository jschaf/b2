package mdext

import (
	"fmt"
	"testing"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/cite/bibtex"
	"github.com/jschaf/b2/pkg/htmls/tags"
)

func newCiteIEEE(key bibtex.Key, order string) string {
	attrs := fmt.Sprintf(`id=%s data-cite-key="%s"`, "cite_"+key, key)
	return tags.CiteAttrs(attrs, order)
}

func TestNewCitationExt_IEEE(t *testing.T) {
	style := cite.IEEE
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"ignores prefix and suffix",
			"[**qux**, @bib_foo *bar*]",
			tags.P(newCiteIEEE("bib_foo", "[1]")),
		},
		{
			"keeps surrounding text",
			"alpha [@bib_foo] bravo",
			tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo"),
		},
		{
			"numbers different citations",
			"alpha [@bib_foo] bravo [@bib_bar]",
			tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo ", newCiteIEEE("bib_bar", "[2]")),
		},
		{
			"re-use citation numbers",
			"alpha [@bib_foo] bravo [@bib_bar] charlie [@bib_foo] delta [@bib_bar]",
			tags.P(
				"alpha ", newCiteIEEE("bib_foo", "[1]"),
				" bravo ", newCiteIEEE("bib_bar", "[2]"),
				" charlie ", newCiteIEEE("bib_foo", "[1]"),
				" delta ", newCiteIEEE("bib_bar", "[2]"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := newMdTester(t, NewCitationExt(style))
			SetTOMLMeta(ctx, PostMeta{
				BibPaths: []string{"./testdata/citation_test.bib"},
			})
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
		})
	}
}
