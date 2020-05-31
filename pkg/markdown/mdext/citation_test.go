package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/htmls/tags"
)

func TestNewCitationExt_IEEE(t *testing.T) {
	style := cite.IEEE
	tests := []struct {
		src  string
		want string
	}{
		{"[@bib_foo *bar*]", tags.P(tags.CiteAttrs(
			`id=cite_bib_foo data-cite-key="bib_foo"`, "[1]"))},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := newMdTester(t, NewCitationExt(style))
			SetTOMLMeta(ctx, PostMeta{
				BibPaths: []string{"./testdata/citation_test.bib"},
			})
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
		})
	}
}
