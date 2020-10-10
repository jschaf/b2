package mdext

import (
	"bytes"
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
