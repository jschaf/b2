package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/texts"
)

func TestNewColonBlockExt(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"h1 + p + colon",
			texts.Dedent(`
        # header
        foo

        ::: preview
        qux
        :::
      `),
			texts.Dedent(`
        <h1>header</h1>
			  <p>foo</p>
      `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := newMdTester(t, NewColonBlockExt())
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
		})
	}
}
