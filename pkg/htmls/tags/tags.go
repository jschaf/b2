package tags

import "strings"

func P(ts ...string) string {
	return Wrap("p", ts...)
}

func Code(ts ...string) string {
	return Wrap("code", ts...)
}

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
