package htmls

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/texts"
	"golang.org/x/net/html"
	"testing"
)

func TestRenderNode(t *testing.T) {
	tests := []struct {
		node *html.Node
		want string
	}{
		{elem("p"), "<p></p>"},
		{elem("p", text("foo")), texts.Dedent("<p>\n  foo\n</p>")},
		{elem("article", elem("p", text("foo"), elem("p", text("bar")))),
			texts.Dedent("<article>\n    <p>\n    foo        <p>\n      bar\n    </p>\n  </p>\n</article>")},
	}
	for _, tt := range tests {
		nameBuf := new(bytes.Buffer)
		_ = html.Render(nameBuf, tt.node)
		name := nameBuf.String()
		t.Run(name, func(t *testing.T) {
			b := new(bytes.Buffer)
			RenderNode(tt.node, b, 0)
			got := b.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("RenderNode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
