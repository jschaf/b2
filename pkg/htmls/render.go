package htmls

import (
	"bytes"
	"golang.org/x/net/html"
	"strings"
)

// RenderNodes prints a normalized string representation of HTML nodes.
func RenderNodes(nodes []*html.Node) string {
	b := new(bytes.Buffer)
	for i, node := range nodes {
		RenderNode(node, b, 0)
		if i < len(nodes)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// RenderNode prints a normalized string representation of an HTML.
func RenderNode(n *html.Node, w *bytes.Buffer, indent int) {
	prefix := strings.Repeat("  ", indent)
	w.WriteString(prefix)
	if n.Type != html.ElementNode {
	}
	switch n.Type {
	case html.ElementNode:
		switch n.Data {
		case "p", "div", "ul", "ol", "article":
			renderBlockElem(n, w, indent)
			return
		}
		_ = html.Render(w, n)
	default:
		_ = html.Render(w, n)
	}

}

func renderBlockElem(n *html.Node, w *bytes.Buffer, indent int) {
	prefix := bytes.Repeat([]byte("  "), indent)
	w.Write(prefix)
	w.WriteByte('<')
	w.WriteString(n.Data)
	for _, a := range n.Attr {
		w.WriteByte(' ')
		if a.Namespace != "" {
			w.WriteString(a.Namespace)
			w.WriteByte(':')
		}
		w.WriteString(a.Key)
		w.WriteString(`="`)
		escape(w, a.Val)
		w.WriteByte('"')
	}

	if voidElements[n.Data] {
		w.WriteString("/>")
		return
	}
	w.WriteByte('>')

	if n.FirstChild != nil {
		w.WriteByte('\n')
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		RenderNode(c, w, indent+1)
	}

	if n.FirstChild != nil {
		w.WriteByte('\n')
	}
	w.Write(prefix)
	w.WriteString("</")
	w.WriteString(n.Data)
	w.WriteByte('>')
}

const escapedChars = "&'<>\"\r"

func escape(w *bytes.Buffer, s string) {
	i := strings.IndexAny(s, escapedChars)
	for i != -1 {
		w.WriteString(s[:i])
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			esc = "&apos;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&quot;"
		case '\r':
			esc = "&#13;"
		default:
			panic("unrecognized escape character")
		}
		s = s[i+1:]
		w.WriteString(esc)
		i = strings.IndexAny(s, escapedChars)
	}
	w.WriteString(s)
}

// Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
var voidElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"keygen": true,
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}
