package mdext

import (
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type TableExt struct {
}

func NewTableExt() TableExt {
	return TableExt{}
}

func (t TableExt) Extend(m goldmark.Markdown) {
	extenders.Extend(m, extension.Table)
}
