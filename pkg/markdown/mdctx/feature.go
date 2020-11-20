package mdctx

type Feature string

const (
	FeatureKatex Feature = "katex"
)

// Features is a special feature of a post. If a post has a feature, we might
// need to change the HTML template, like load custom CSS.
type Features struct {
	feats map[Feature]struct{}
}

func NewFeatures() *Features {
	return &Features{feats: make(map[Feature]struct{})}
}

func (fs *Features) Add(f Feature) {
	fs.feats[f] = struct{}{}
}

func (fs *Features) AddAll(f2 *Features) {
	for f := range f2.feats {
		fs.feats[f] = struct{}{}
	}
}

func (fs *Features) Has(f Feature) bool {
	_, ok := fs.feats[f]
	return ok
}

func (fs *Features) Len() int {
	return len(fs.feats)
}

func (fs *Features) Slice() []Feature {
	s := make([]Feature, 0, len(fs.feats))
	for f := range fs.feats {
		s = append(s, f)
	}
	return s
}
