package attrs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark/ast"
)

func TestAddClass(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		classes  []string
		want     string
	}{
		{"no class", "", []string{"foo", "bar", "baz"}, "foo bar baz"},
		{"existing class", "qux bar", []string{"foo", "bar", "baz"}, "qux bar foo bar baz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := ast.NewCodeSpan()
			if tt.existing != "" {
				n.SetAttribute([]byte("class"), []byte(tt.existing))
			}
			AddClass(n, tt.classes...)

			got, ok := n.Attribute([]byte("class"))
			if !ok {
				t.Errorf("class attribute not found on node")
			}

			if diff := cmp.Diff(tt.want, string(got.([]byte))); diff != "" {
				t.Errorf("Class attribute mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
