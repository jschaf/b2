package mdctx

type Feature string

const (
	FeatureKatex Feature = "katex"
)

// FeatureSet is a special feature of a post. If a post has a feature, we might
// need to change the HTML template, like load custom CSS.
type FeatureSet struct {
	feats map[Feature]struct{}
}

func NewFeatureSet() *FeatureSet {
	return &FeatureSet{feats: make(map[Feature]struct{})}
}

func (fs *FeatureSet) Add(f Feature) {
	fs.feats[f] = struct{}{}
}

func (fs *FeatureSet) AddAll(f2 *FeatureSet) {
	for f := range f2.feats {
		fs.feats[f] = struct{}{}
	}
}

func (fs *FeatureSet) Has(f Feature) bool {
	_, ok := fs.feats[f]
	return ok
}

func (fs *FeatureSet) Len() int {
	return len(fs.feats)
}

func (fs *FeatureSet) Slice() []Feature {
	s := make([]Feature, 0, len(fs.feats))
	for f := range fs.feats {
		s = append(s, f)
	}
	return s
}
