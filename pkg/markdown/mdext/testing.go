package mdext

import (
	"bytes"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/htmls"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.uber.org/zap/zaptest"
)

const testPostPath = "/md/test/path.md"

// newMdTester creates a new markdown with the given extensions. We can't use
// our top level markdown because it would create a circular dependency.
func newMdTester(t *testing.T, exts ...goldmark.Extender) (goldmark.Markdown, parser.Context) {
	md := goldmark.New(goldmark.WithExtensions(exts...))
	pc := parser.NewContext()
	logger := zaptest.NewLogger(t)
	SetFilePath(pc, testPostPath)
	SetLogger(pc, logger)
	SetRenderer(pc, md.Renderer())

	return md, pc
}

// mustParseMarkdown parses markdown into a document node.
func mustParseMarkdown(t *testing.T, md goldmark.Markdown, ctx parser.Context, src string) ast.Node {
	t.Helper()
	reader := text.NewReader([]byte(src))
	doc := md.Parser().Parse(reader, parser.WithContext(ctx))
	if errs := PopErrors(ctx); len(errs) > 0 {
		t.Fatalf("errors while parsing: %v", errs)
	}
	return doc
}

// assertNoRenderDiff asserts the markdown instance renders src into the wanted
// string.
func assertNoRenderDiff(t *testing.T, doc ast.Node, md goldmark.Markdown, src, want string) {
	t.Helper()
	bufW := &bytes.Buffer{}
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
