package mdext

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/htmls"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"

	"golang.org/x/net/html"
)

func TestArticleTransformer_Transform(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"h1 + p",
			texts.Dedent(`
				# header
				foo

				bar
      `),
			texts.Dedent(`
          <article>
					<time datetime="0001-01-01T00:00:00Z">January  1, 0001</time>
					<a href="/" title="header"><h1>header</h1>
					</a><p>foo</p>
					</article>
					<p>bar</p>`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewArticleExt(),
				NewTimeExt(),
			))
			buf := new(bytes.Buffer)
			ctx := parser.NewContext()
			if err := md.Convert([]byte(tt.src), buf, parser.WithContext(ctx)); err != nil {
				t.Fatal(err)
			}

			if diff, err := htmls.Diff(buf, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			} else if diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func assertHTMLEqual(t *testing.T, buf io.Reader, expected string) {
	testCtx := &html.Node{Type: html.ElementNode, Data: "test"}
	gotR, _ := html.ParseFragment(buf, testCtx)
	wantR, _ := html.ParseFragment(strings.NewReader(expected), testCtx)
	got := normalizeNodes(gotR)
	want := normalizeNodes(wantR)
	fmt.Println("!! want")
	for _, node := range want {
		dumpNode(node, 0)
	}
	fmt.Println("!! got")
	for _, node := range got {
		dumpNode(node, 0)
	}
	if !reflect.DeepEqual(got, want) {
		gotW := new(bytes.Buffer)
		for _, g := range got {
			_ = html.Render(gotW, g)
		}

		wantW := new(bytes.Buffer)
		for _, w := range want {
			_ = html.Render(wantW, w)
		}
		t.Errorf("want:\n%s\ngot:\n%s", wantW, gotW)
	}
}

func normalizeNodes(nodes []*html.Node) []*html.Node {
	ns := make([]*html.Node, 0, len(nodes))
	for _, node := range nodes {
		d := node.Data
		fmt.Println(d)
		if isEmptyNode(node) {
			continue
		}
		normalizeNode(node)
		ns = append(ns, node)
	}
	return ns
}

func normalizeNode(node *html.Node) {
	d := node.Data
	fmt.Println("  normalize node: " + d)

	switch node.Type {
	case html.TextNode:
		fmt.Println("  got text node: " + node.Data)
		if isEmptyNode(node) {
			if node.Parent == nil {
				return
			}
			node.Parent.RemoveChild(node)
		}
	case html.ElementNode:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			normalizeNode(c)
		}
	}

}

func isEmptyNode(node *html.Node) bool {
	return strings.TrimSpace(node.Data) == ""
}

func dumpNode(node *html.Node, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch node.Type {
	case html.ElementNode:
		tag := node.Data
		fmt.Printf("%s%s", prefix, tag)
		fc := node.FirstChild
		if fc == nil {
			fmt.Printf(" {}\n")
			return
		}

		fmt.Printf(" {\n")
		for c := fc; c != nil; c = c.NextSibling {
			dumpNode(c, indent+1)
		}
		fmt.Printf("%s}\n", prefix)
	case html.TextNode:
		hi := 30
		if hi > len(node.Data) {
			hi = len(node.Data)
		}
		fmt.Printf("%sText {'%s'}\n", prefix, strings.Replace(node.Data[:hi], "\n", "\\n", -1))
	}
}
