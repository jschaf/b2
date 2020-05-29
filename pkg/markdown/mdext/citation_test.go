package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/htmls/tags"
)

func TestNewCitationExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"[@joe *bar*]", tags.P(tags.Cite("@joe ", tags.Em("bar")))},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := newMdTester(t, NewCitationExt())
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
		})
	}
}
