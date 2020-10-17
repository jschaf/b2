package mdext

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// FootnoteExt is the Goldmark extension to render a markdown footnote.
type FootnoteExt struct{}

func (f *FootnoteExt) Extend(m goldmark.Markdown) {
	extension.Footnote.Extend(m)
}

func NewFootnoteExt() *FootnoteExt {
	return &FootnoteExt{}
}
