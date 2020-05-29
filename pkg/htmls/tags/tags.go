package tags

import "strings"

func Wrap(tag string, contents ...string) string {
	startTagSize := len(tag) + 2
	endTagSize := startTagSize + 1
	size := startTagSize + endTagSize
	for _, content := range contents {
		size += len(content)
	}
	b := strings.Builder{}
	b.Grow(size)

	b.WriteString("<" + tag + ">")
	for _, t := range contents {
		b.WriteString(t)
	}
	b.WriteString("</" + tag + ">")
	return b.String()
}

func Cite(ts ...string) string {
	return Wrap("cite", ts...)
}

func Em(ts ...string) string {
	return Wrap("em", ts...)
}

func Code(ts ...string) string {
	return Wrap("code", ts...)
}

func P(ts ...string) string {
	return Wrap("p", ts...)
}
