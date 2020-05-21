package mdext

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/htmls"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
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
	buf := &bytes.Buffer{}
	if err := md.Convert([]byte(src), buf, parser.WithContext(ctx)); err != nil {
		t.Fatal(err)
	}

	if diff, err := htmls.Diff(buf, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	} else if diff != "" {
		t.Errorf("Render mismatch (-want +got):\n%s", diff)
	}
}

// assertCtxContainsAll asserts that the content contains every wanted
// key-value pair.
func assertCtxContainsAll(t *testing.T, got parser.Context, want map[parser.ContextKey]interface{}) {
	t.Helper()

	for k, v := range want {
		if got := got.Get(k); !reflect.DeepEqual(got, v) {
			t.Errorf("context key %v, got %s, want %v", k, got, want[k])
		}
	}
}
