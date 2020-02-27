package mdext

import "github.com/yuin/goldmark/parser"

var assetsCtxKey = parser.NewContextKey()

// GetAssets returns a map of all assets associated with a post.
// A map of the relative URL to the full file path of an asset like an image.
// For example, 1 entry might be ./img.png -> /home/joe/blog/img.png.
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
