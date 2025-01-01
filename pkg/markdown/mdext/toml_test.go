package mdext

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jschaf/jsc/pkg/git"
	"github.com/jschaf/jsc/pkg/markdown/mdtest"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/jsc/pkg/texts"
)

func TestMeta(t *testing.T) {
	root := git.RootDir()
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
        bib_paths = ["./ref.bib", "/r1/r2.bib"]
				+++
				# Hello goldmark-meta
      `),
			texts.Dedent(`
        <h1>Hello goldmark-meta</h1>
      `),
			PostMeta{
				Path:     "/a_slug/",
				Slug:     "a_slug",
				Date:     time.Date(2019, time.September, 20, 0, 0, 0, 0, time.Local),
				BibPaths: []string{"/md/test/ref.bib", filepath.Join(root, "r1/r2.bib")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewTOMLExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
			if diff := cmp.Diff(GetTOMLMeta(ctx), tt.wantTOMLMeta, cmpTimeDate()); diff != "" {
				t.Errorf("TOML meta mismatch: (-got +want)\n%s", diff)
			}
		})
	}
}

func cmpTimeDate() cmp.Option {
	return cmp.Transformer("TimeDate", func(t time.Time) string {
		return t.Format(time.DateOnly)
	})
}
