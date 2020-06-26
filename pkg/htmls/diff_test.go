package htmls

import (
	"github.com/go-test/deep"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestDiffStrings(t *testing.T) {
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
			diff, err := DiffStrings(tt.x, tt.y)
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

func Test_parseFragment(t *testing.T) {
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
			got, err := parseFragment(strings.NewReader(tt.input))
			if err != nil {
				t.Errorf("parseFragment() error = %v", err)
				return
			}
			gotR, wantR := RenderNodes(got), RenderNodes(tt.want)
			if gotR != wantR {
				t.Errorf("got:\n%s\nwant:\n%s", gotR, wantR)

			}
			if diff := deep.Equal(gotR, wantR); diff != nil {
				t.Error(diff)
			}
		})
	}
}
