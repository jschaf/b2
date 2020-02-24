package mdext

import "github.com/yuin/goldmark/parser"

var assetsCtxKey = parser.NewContextKey()

func GetAssets(pc parser.Context) map[string]string {
	m := pc.Get(assetsCtxKey)
	if _, ok := m.(map[string]string); m == nil || !ok {
		m = make(map[string]string)
		pc.Set(assetsCtxKey, m)
	}

	return m.(map[string]string)
}

func AddAsset(pc parser.Context, path, src string) {
	m := pc.Get(assetsCtxKey)
	if _, ok := m.(map[string]string); m == nil || !ok {
		m = make(map[string]string)
		pc.Set(assetsCtxKey, m)
	}

	x := m.(map[string]string)
	x[path] = src
}
