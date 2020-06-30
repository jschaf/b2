package attrs

import (
	"github.com/yuin/goldmark/ast"
)

// AddClass adds additional classes to the node attributes, preserving existing
// class attributes.
func AddClass(n ast.Node, class ...string) {
	size := 0
	for _, c := range class {
		size += len(c)
	}
	size += len(class) // spaces in between old

	var old []byte
	raw, ok := n.Attribute([]byte("class"))
	if ok {
		old = raw.([]byte)
	}

	newer := old
	if cap(old) < len(old)+size {
		// The existing slice capacity can't hold everything so allocate a new slice.
		newer = make([]byte, len(old), len(old)+size)
		copy(newer, old)
	}

	if len(newer) > 0 {
		newer = append(newer, ' ')
	}

	for i, cls := range class {
		newer = append(newer, cls...)
		if i < len(class)-1 {
			newer = append(newer, ' ')
		}
	}
	n.SetAttribute([]byte("class"), newer)
}

func GetStringAttr(n ast.Node, k string) string {
	a, ok := n.AttributeString(k)
	if !ok {
		return ""
	}
	switch s := a.(type) {
	case []byte:
		return string(s)
	case string:
		return s
	default:
		return ""
	}
}
