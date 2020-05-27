package mdext

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/htmls"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.uber.org/zap/zaptest"
)

// newMdTester creates a new markdown with the given extensions. We can't use
// our top level markdown because it would create a circular dependency.
func newMdTester(t *testing.T, exts ...goldmark.Extender) (goldmark.Markdown, parser.Context) {
	md := goldmark.New(goldmark.WithExtensions(exts...))
	pc := parser.NewContext()
	logger := zaptest.NewLogger(t)
	SetFilePath(pc, "<md_test_file>")
	SetLogger(pc, logger)
	SetRenderer(pc, md.Renderer())

	return md, pc
}

// assertNoRenderDiff asserts the markdown instance renders src into the wanted
// string.
func assertNoRenderDiff(t *testing.T, md goldmark.Markdown, ctx parser.Context, src, want string) {
	t.Helper()
	bufW := &bytes.Buffer{}
	reader := text.NewReader([]byte(src))
	doc := md.Parser().Parse(reader, parser.WithContext(ctx))
	if testing.Verbose() {
		doc.Dump([]byte(src), 0)
	}

	if err := md.Renderer().Render(bufW, []byte(src), doc); err != nil {
		t.Fatal(err)
	}

	if diff, err := htmls.Diff(bufW, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	} else if diff != "" {
		t.Errorf("Render mismatch (-want +got):\n%s", diff)
	}
}

// assertCtxContainsAll asserts that the content contains every wanted
// key-value pair.
func assertCtxContainsAll(t *testing.T, got parser.Context, wants map[parser.ContextKey]interface{}) {
	t.Helper()

	for key, want := range wants {
		got := got.Get(key)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Context key mismatch for key=%d (-want +got):\n%s", key, diff)
		}
	}
}
