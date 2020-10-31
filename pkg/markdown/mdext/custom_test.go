package mdext

import (
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/yuin/goldmark/ast"
	"testing"
)

func TestNewCustomExt(t *testing.T) {
	tests := []struct {
		name string
		src  ast.Node
		want string
	}{
		{
			"custom cite",
			(func() ast.Node {
				t := NewCustomInline("cite")
				s := ast.NewString([]byte("blah"))
				t.AppendChild(t, s)
				t.SetAttributeString("foo", "bar")
				return t
			})(),
			`<cite foo="bar">blah</cite>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, _ := mdtest.NewTester(t, NewCustomExt())
			mdtest.AssertNoRenderDiff(t, tt.src, md, "", tt.want)
		})
	}
}
