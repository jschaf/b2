package mdext

import (
	"fmt"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type FnAction int

const (
	FnOrderNext FnAction = iota // the given order was used so advance to next order
	FnOrderKeep                 // the given order was not used so keep the current order
)

// FootnoteOrderer is implemented by all nodes that need a global, monotonic
// ordering. The use case is we want side notes and citations to not reuse the
// same number.
//
// Applies to all node kinds that share a global ordering to
// synchronize the display numbering of footnotes like [1], [2], [3]. Since
// multiple nodes might be ordered, we'll use a transformer. We could put this
// in parser.Context and apply the ordering during parsing except that citations
// are "parsed" in a transformer because Goldmark hijacks link parsing.
type FootnoteOrderer interface {
	// FootnoteOrder gives a node the nextOrder and all previous orders in seen.
	// The node should return whether the nextOrder was used or a previous order
	// was used, as well as the node ID as a string.
	FootnoteOrder(nextOrder int, seen map[string]int, pc parser.Context) (FnAction, string)
}

type footnoteOrderTransformer struct {
}

func (f footnoteOrderTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	num := 1
	seen := make(map[string]int)
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		orderer, ok := n.(FootnoteOrderer)
		if !ok {
			return ast.WalkContinue, nil
		}
		action, key := orderer.FootnoteOrder(num, seen, pc)
		switch action {
		case FnOrderNext:
			seen[key] = num
			num++
		case FnOrderKeep: // do nothing
		default:
			return ast.WalkStop, fmt.Errorf("unhandled footnote order action: %d", action)
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		mdctx.PushError(pc, fmt.Errorf("walk footnote order transformer: %w", err))
		return
	}
}
