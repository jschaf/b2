package mdext

import (
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestWalkStop(t *testing.T) {
	d := ast.NewDocument()
	h1 := ast.NewHeading(1)
	h2 := ast.NewHeading(2)
	d.AppendChild(d, h1)
	d.AppendChild(d, h2)

	var firstH *ast.Heading
	err := ast.Walk(d, func(n ast.Node, _ bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindHeading {

			firstH = n.(*ast.Heading)
			return ast.WalkStop, nil

		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if firstH.Level != h1.Level {
		t.Errorf("expected level %d, got level %d", h1.Level, firstH.Level)
	}
}
