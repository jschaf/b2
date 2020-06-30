package asts

import (
	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"testing"
)

func TestWriteSlugText(t *testing.T) {
	tests := []struct {
		src  string
		size int
		want string
	}{
		{"# h1 *em*", 32, "h1-em"},
		{"# h1 - _ *em*", 32, "h1-em"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			b := make([]byte, tt.size)
			md, ctx := mdtest.NewTester(t)
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			n := doc.FirstChild()
			got := string(WriteSlugText(b, n, []byte(tt.src)))
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("WriteSlugText() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
