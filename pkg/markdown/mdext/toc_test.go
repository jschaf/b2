package mdext

import (
	"github.com/jschaf/b2/pkg/texts"
	"testing"
)

func TestNewTOCExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{
			texts.Dedent(`
       :toc:
			
       # title
     `),
			texts.Dedent(`
			   <div class="toc"></div>
				 <h1>title</h1>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := newMdTester(t, NewColonLineExt(), NewTOCExt())
			doc := mustParseMarkdown(t, md, ctx, tt.src)
			assertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
