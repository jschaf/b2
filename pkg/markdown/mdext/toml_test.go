package mdext

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/texts"
)

func TestMeta(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		want         string
		wantTOMLMeta PostMeta
	}{
		{
			"slug date + h1",
			texts.Dedent(`
				+++
				slug = "a_slug"
				date = 2019-09-20
				+++
				# Hello goldmark-meta
      `),
			texts.Dedent(`
        <h1>Hello goldmark-meta</h1>
      `),
			PostMeta{
				Path: "/a_slug",
				Slug: "a_slug",
				Date: time.Date(2019, time.September, 20, 0, 0, 0, 0, time.Local),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := newMdTester(t, NewTOMLExt())
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
			if diff := cmp.Diff(GetTOMLMeta(ctx), tt.wantTOMLMeta); diff != "" {
				t.Errorf("TOML meta mismatch: (-got +want)\n%s", diff)
			}
		})
	}
}
