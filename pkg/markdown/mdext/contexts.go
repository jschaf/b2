package mdext

import (
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"go.uber.org/zap"
)

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

var titleCtxKey = parser.NewContextKey()

// Returns the title as parsed from the first H1 header in the document.
func GetTitle(pc parser.Context) string {
	return pc.Get(titleCtxKey).(string)
}

func SetTitle(pc parser.Context, title string) {
	pc.Set(titleCtxKey, title)
}

var filePathCtxKey = parser.NewContextKey()

// GetFilePath returns the file path of the main markdown file that initiated
// this parse. Helpful for error messages.
func GetFilePath(pc parser.Context) string {
	p := pc.Get(filePathCtxKey)
	if p == nil {
		return ""
	}
	return p.(string)
}

// SetFilePath sets the file path of the main markdown path. Should only be
// called once per file.
func SetFilePath(pc parser.Context, val string) {
	pc.Set(filePathCtxKey, val)
}

var previewCtxKey = parser.NewContextKey()

// AddPreview adds a preview to the context so that it can be rendered into
// the corresponding link.
func AddPreview(pc parser.Context, p Preview) {
	if existing := pc.Get(previewCtxKey); existing == nil {
		pc.Set(previewCtxKey, make(map[string]Preview))
	}

	previews := pc.Get(previewCtxKey).(map[string]Preview)
	previews[p.URL] = p
}

// GetPreview returns the preview, if any, for the URL. Returns an empty Preview
// and false if no preview exists for the URL.
func GetPreview(pc parser.Context, url string) (Preview, bool) {
	previews, ok := pc.Get(previewCtxKey).(map[string]Preview)
	if !ok {
		return Preview{}, false
	}
	p, ok := previews[url]
	return p, ok
}

var rendererCtxKey = parser.NewContextKey()

// GetRenderer returns the main goldmark renderer or nil if none exists. Useful
// for rendering markdown into HTML that doesn't fit into the
// parse-transform-render model. For example, link preview data is rendered into
// HTML stored in the data attributes of <a> tags, like:
//   <a data-preview-title="<h1>foo</h1>">
func GetRenderer(pc parser.Context) (renderer.Renderer, bool) {
	r := pc.Get(rendererCtxKey)
	if r == nil {
		return nil, false
	}
	return r.(renderer.Renderer), true
}

// SetRenderer sets the goldmark renderer used to render markdown. Should only
// be called once per file.
func SetRenderer(pc parser.Context, r renderer.Renderer) {
	pc.Set(rendererCtxKey, r)
}

var loggerCtxKey = parser.NewContextKey()

func GetLogger(pc parser.Context) *zap.Logger {
	return pc.Get(loggerCtxKey).(*zap.Logger)
}

func SetLogger(pc parser.Context, l *zap.Logger) {
	pc.Set(loggerCtxKey, l)
}
