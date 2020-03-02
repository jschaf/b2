package mdext

import "github.com/yuin/goldmark/parser"

var filePathCtxKey = parser.NewContextKey()

func GetFilePath(pc parser.Context) string {
	p := pc.Get(filePathCtxKey)
	if p == nil {
		return ""
	}
	return p.(string)
}

func SetFilePath(pc parser.Context, val string) {
	pc.Set(filePathCtxKey, val)
}
