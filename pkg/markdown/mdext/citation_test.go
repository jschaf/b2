package mdext

import (
	"testing"
)

func TestNewCitationExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"qux [foo] bar\n\n[foo]: http://example.com", "<p>foo</p>"},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := newMdTester(t, NewCitationExt())
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
		})
	}
}
