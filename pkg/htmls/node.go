package htmls

import "golang.org/x/net/html"

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
