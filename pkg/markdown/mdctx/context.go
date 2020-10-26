package mdctx

import (
	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"go.uber.org/zap"
)

var errorsCtxKey = parser.NewContextKey()

func PushError(pc parser.Context, err error) {
	var errs []error
	if e := pc.Get(errorsCtxKey); e != nil {
		errs = e.([]error)
	}
	errs = append(errs, err)
	pc.Set(errorsCtxKey, errs)
}

func PopErrors(pc parser.Context) []error {
	var errs []error
	if e := pc.Get(errorsCtxKey); e != nil {
		errs = e.([]error)
	}
	pc.Set(errorsCtxKey, nil)
	return errs
}

var AssetsCtxKey = parser.NewContextKey()

// GetAssets returns a map of all assets associated with a post.
// A map of the relative URL to the full file path of an asset like an image.
// For example, 1 entry might be ./img.png -> /home/joe/blog/img.png.
func GetAssets(pc parser.Context) assets.Map {
	m := pc.Get(AssetsCtxKey)
	if _, ok := m.(map[string]string); m == nil || !ok {
		m = make(map[string]string)
		pc.Set(AssetsCtxKey, m)
	}
	return m.(map[string]string)
}

func AddAsset(pc parser.Context, path, src string) {
	m := pc.Get(AssetsCtxKey)
	if _, ok := m.(assets.Map); m == nil || !ok {
		m = make(assets.Map)
		pc.Set(AssetsCtxKey, m)
	}

	x := m.(assets.Map)
	x[path] = src
}

var featuresCtxKey = parser.NewContextKey()

func GetFeatures(pc parser.Context) *Features {
	fs := pc.Get(featuresCtxKey)
	if _, ok := fs.(*Features); fs == nil || !ok {
		fs = NewFeatures()
		pc.Set(featuresCtxKey, fs)
	}
	return fs.(*Features)
}

func AddFeature(pc parser.Context, feat Feature) {
	fs := pc.Get(featuresCtxKey)
	if _, ok := fs.(*Features); fs == nil || !ok {
		fs = NewFeatures()
	}
	fs.(*Features).Add(feat)
	pc.Set(featuresCtxKey, fs)
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

var headingIDsCtxKey = parser.NewContextKey()

func HeadingIDs(pc parser.Context) map[string]struct{} {
	rawIDs := pc.Get(headingIDsCtxKey)
	if rawIDs == nil {
		rawIDs = make(map[string]struct{})
		pc.Set(headingIDsCtxKey, rawIDs)
	}
	return rawIDs.(map[string]struct{})
}
