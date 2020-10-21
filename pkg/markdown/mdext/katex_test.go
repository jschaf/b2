package mdext

import (
	"bytes"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"strings"
	"testing"
)

func TestNewKatexExt_works(t *testing.T) {
	md, ctx := mdtest.NewTester(t, NewKatexExt())
	src := "$a=1$"
	doc := mdtest.MustParseMarkdown(t, md, ctx, src)
	buf := &bytes.Buffer{}
	if err := md.Renderer().Render(buf, []byte(src), doc); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "katex-mathml") {
		t.Errorf("expected %s to render to include katex-mathml; but got:\n%s", src, buf.String())
	}
}

func TestNewKatexExt_withMath_addsKatexFeature(t *testing.T) {
	md, ctx := mdtest.NewTester(t, NewKatexExt())
	src := "$a=1$"
	doc := mdtest.MustParseMarkdown(t, md, ctx, src)
	buf := &bytes.Buffer{}
	if err := md.Renderer().Render(buf, []byte(src), doc); err != nil {
		t.Fatal(err)
	}
	feats := mdctx.GetFeatures(ctx)
	for _, feat := range feats.Slice() {
		if feat == mdctx.FeatureKatex {
			return
		}
	}
	t.Fatalf("expected katex feature in context features %v; but was missing", feats)
}

func TestNewKatexExt_withoutMath_doesntAddKatexFeature(t *testing.T) {
	md, ctx := mdtest.NewTester(t, NewKatexExt())
	src := "# Price is $10"
	doc := mdtest.MustParseMarkdown(t, md, ctx, src)
	buf := &bytes.Buffer{}
	if err := md.Renderer().Render(buf, []byte(src), doc); err != nil {
		t.Fatal(err)
	}
	feats := mdctx.GetFeatures(ctx)
	for _, feat := range feats.Slice() {
		if feat == mdctx.FeatureKatex {
			t.Fatalf("expected no katex feature in context features %v", feats)
		}
	}
}
