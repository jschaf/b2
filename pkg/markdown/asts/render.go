package asts

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

func NopRender(util.BufWriter, []byte, ast.Node, bool) (ast.WalkStatus, error) {
	return ast.WalkSkipChildren, nil
}
