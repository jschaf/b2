package htmls

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// ParseFragment parses a normalize version of an HTML node.
func ParseFragment(r io.Reader) ([]*html.Node, error) {
	testCtx := &html.Node{Type: html.ElementNode, Data: "normalizedFragment"}
	nodes, err := html.ParseFragment(r, testCtx)
	if err != nil {
		return nil, err
	}
	return normalizeNodes(nodes), nil
}

// Diff returns the diff between the normalized HTML fragments.
func Diff(x, y io.Reader) (string, error) {
	frag1, err := ParseFragment(x)
	if err != nil {
		return "", err
	}

	frag2, err := ParseFragment(y)
	if err != nil {
		return "", err
	}

	dump1 := DumpNodes(frag1)
	dump2 := DumpNodes(frag2)
	if dump1 == dump2 {
		return "", nil
	}
	diff := fmt.Sprintf("got:\n%s\nwant:\n%s", dump1, dump2)
	return diff, nil
}

func normalizeNodes(nodes []*html.Node) []*html.Node {
	ns := make([]*html.Node, 0, len(nodes))
	for _, node := range nodes {
		if isEmptyNode(node) {
			continue
		}
		normalizeNode(node)
		ns = append(ns, node)
	}
	return ns
}

func normalizeNode(node *html.Node) {
	switch node.Type {
	case html.TextNode:
		node.Data = strings.TrimSpace(node.Data)
		if isEmptyNode(node) {
			if node.Parent == nil {
				return
			}

			p := node.Parent
			p.RemoveChild(node)
		}
	case html.ElementNode:
		cur := node.FirstChild
		for cur != nil {
			next := cur.NextSibling
			normalizeNode(cur)
			cur = next
		}
	}
}

func isEmptyNode(node *html.Node) bool {
	return strings.TrimSpace(node.Data) == ""
}

// DumpNodes prints a string representation of HTML nodes.
func DumpNodes(nodes []*html.Node) string {
	b := new(bytes.Buffer)
	for _, node := range nodes {
		dumpNode(node, b, 0)
	}
	return b.String()
}

func dumpNode(node *html.Node, buf *bytes.Buffer, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch node.Type {
	case html.ElementNode:
		tag := node.Data
		buf.WriteString(prefix)
		buf.WriteString(tag)
		fc := node.FirstChild
		if fc == nil {
			buf.WriteString(" {}\n")
			return
		}

		buf.WriteString(" {\n")
		for c := fc; c != nil; c = c.NextSibling {
			dumpNode(c, buf, indent+1)
		}
		fmt.Fprintf(buf, "%s}\n", prefix)
	case html.TextNode:
		hi := 30
		if hi > len(node.Data) {
			hi = len(node.Data)
		}
		fmt.Fprintf(buf, "%sText {'%s'}\n", prefix, strings.Replace(node.Data[:hi], "\n", "\\n", -1))
	}
}
