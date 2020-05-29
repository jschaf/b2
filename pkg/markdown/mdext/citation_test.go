package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/htmls/tags"
)

func TestNewCitationExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"[@bib_foo *bar*]", tags.P(tags.CiteAttrs(`data-cite-key="joe"`, "@joe ", tags.Em("bar")))},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := newMdTester(t, NewCitationExt(cite.IEEE))
			SetTOMLMeta(ctx, PostMeta{
				BibPaths: []string{"./testdata/citation_test.bib"},
			})
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
		})
	}
}
