package mdext

import (
	"bytes"
	"github.com/jschaf/b2/pkg/htmls"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
	"testing"
)

func TestCloneNode(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			"h1 + p + p",
			texts.Dedent(`
				# header
				foo

				bar`),
		},
		{
			"h1-em + p link + ul",
			texts.Dedent(`
				# *header*
				foo [bar]
			  
			  [bar]: www.example.com

				- one
			  - two`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewArticleExt(),
				NewTimeExt(),
				NewHeaderExt())
			orig := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			clone := CloneNode(orig)

			origB := &bytes.Buffer{}
			if err := md.Renderer().Render(origB, []byte(tt.src), orig); err != nil {
				t.Fatal(err)
			}

			cloneB := &bytes.Buffer{}
			if err := md.Renderer().Render(cloneB, []byte(tt.src), clone); err != nil {
				t.Fatal(err)
			}

			if diff, err := htmls.Diff(origB, cloneB); err != nil {
				t.Fatal(err)
			} else if diff != "" {
				t.Errorf("CloneNode() render mismatch (-orig +clone):\n%s", diff)
			}
		})
	}
}
