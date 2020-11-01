package tags

import "strings"

func Join(ts ...string) string {
	return strings.Join(ts, "\n")
}

func Attrs(as ...string) string {
	return strings.Join(as, " ")
}

func WrapAttrs(tag string, attrs string, contents ...string) string {
	startTagSize := len(tag) + 2
	endTagSize := startTagSize + 1
	attrsSize := 0
	if len(attrs) > 0 {
		attrsSize = len(attrs) + 1
	}
	size := startTagSize + endTagSize + attrsSize
	for _, content := range contents {
		size += len(content)
	}
	b := strings.Builder{}
	b.Grow(size)

	b.WriteString("<" + tag)
	if len(attrs) > 0 {
		b.WriteString(" ")
		b.WriteString(attrs)
	}
	b.WriteString(">")
	for _, t := range contents {
		b.WriteString(t)
	}
	b.WriteString("</" + tag + ">")
	return b.String()
}

func Wrap(tag string, contents ...string) string {
	return WrapAttrs(tag, "", contents...)
}

func AAttrs(attrs string, ts ...string) string {
	return WrapAttrs("a", attrs, ts...)
}

func AsideAttrs(attrs string, ts ...string) string {
	return WrapAttrs("aside", attrs, ts...)
}

func Cite(ts ...string) string {
	return Wrap("cite", ts...)
}

func CiteAttrs(attrs string, ts ...string) string {
	return WrapAttrs("cite", attrs, ts...)
}

func Code(ts ...string) string {
	return Wrap("code", ts...)
}

func DivAttrs(attrs string, ts ...string) string {
	return WrapAttrs("div", attrs, ts...)
}

func Em(ts ...string) string {
	return Wrap("em", ts...)
}

func EmAttrs(attrs string, ts ...string) string {
	return WrapAttrs("em", attrs, ts...)
}

func H1Attrs(attrs string, ts ...string) string {
	return WrapAttrs("h1", attrs, ts...)
}

func H2(ts ...string) string {
	return Wrap("h2", ts...)
}

func H2Attrs(attrs string, ts ...string) string {
	return WrapAttrs("h2", attrs, ts...)
}

func H3Attrs(attrs string, ts ...string) string {
	return WrapAttrs("h3", attrs, ts...)
}

func P(ts ...string) string {
	return Wrap("p", ts...)
}

func OlAttrs(attrs string, ts ...string) string {
	lis := make([]string, len(ts))
	for i, t := range ts {
		lis[i] = Wrap("li", t)
	}
	return WrapAttrs("ol", attrs, lis...)
}

func SC(ts ...string) string {
	return WrapAttrs("span", "class=small-caps", ts...)
}

func SpanAttrs(attrs string, ts ...string) string {
	return WrapAttrs("span", attrs, ts...)
}

func Strong(ts ...string) string {
	return Wrap("strong", ts...)
}
