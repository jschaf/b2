package mdext

import "github.com/yuin/goldmark/parser"

var pathCtxKey = parser.NewContextKey()

func GetPath(pc parser.Context) string {
	p := pc.Get(pathCtxKey)
	if p == nil {
		return ""
	}
	return p.(string)
}

func SetPath(pc parser.Context, val string) {
	pc.Set(pathCtxKey, val)
}
