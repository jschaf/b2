package htmls

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	"golang.org/x/net/html"
)

func TestParseFragment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*html.Node
	}{
		{"p", "<p>foo</p>", []*html.Node{elem("p", text("foo"))}},
		{
			"p with space",
			"\n<p>\nfoo</p>\n   ",
			[]*html.Node{elem("p", text("foo"))},
		},
		{
			"div > p with space",
			"<div>\n<p>\nfoo</p>\n   </div>",
			[]*html.Node{elem("div", elem("p", text("foo")))},
		},
		{
			"span + div > p with space",
			"<span>qux   </span>\n  <div>\n<p>\nfoo</p>\n   </div>",
			[]*html.Node{
				elem("span", text("qux")),
				elem("div", elem("p", text("foo"))),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFragment(strings.NewReader(tt.input))
			if err != nil {
				t.Errorf("ParseFragment() error = %v", err)
				return
			}
			gotR, wantR := DumpNodes(got), DumpNodes(tt.want)
			if gotR != wantR {
				t.Errorf("got:\n%s\nwant:\n%s", gotR, wantR)

			}
			if diff := deep.Equal(gotR, wantR); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func elem(tag string, children ...*html.Node) *html.Node {
	node := &html.Node{
		Type: html.ElementNode,
		Data: tag,
	}
	for _, child := range children {
		node.AppendChild(child)
	}
	return node
}

func text(data string) *html.Node {
	return &html.Node{
		Type: html.TextNode,
		Data: data,
	}
}

func TestDiff(t *testing.T) {
	tests := []struct {
		x, y   string
		isSame bool
	}{
		{"<p>foo</p>", "<p>foo</p>", true},
		{"\n<p>\n  foo\n</p>  ", "<p>foo</p>", true},
		{"<p><div>foo</div></p>", "<p>  <div>  foo  </div>  </p>", true},
		{"<p>foo</p>", "<p>bar</p>", false},
	}
	for _, tt := range tests {
		t.Run(tt.x, func(t *testing.T) {
			diff, err := Diff(strings.NewReader(tt.x), strings.NewReader(tt.y))
			if err != nil {
				t.Errorf("Diff() error = %v", err)
				return
			}
			if tt.isSame && diff != "" {
				t.Errorf("Diff() got = %v, want no diff", diff)
			}
		})
	}
}
