package mdtest

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/htmls"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.uber.org/zap/zaptest"
	"strings"
	"testing"
)

const PostPath = "/md/test/path.md"

// NewTester creates a new markdown with the given extensions. We can't use
// our top level markdown because it would create a circular dependency.
func NewTester(t *testing.T, exts ...goldmark.Extender) (goldmark.Markdown, parser.Context) {
	md := goldmark.New(goldmark.WithExtensions(exts...))
	pc := parser.NewContext()
	logger := zaptest.NewLogger(t)
	mdctx.SetFilePath(pc, PostPath)
	mdctx.SetLogger(pc, logger)
	mdctx.SetRenderer(pc, md.Renderer())

	return md, pc
}

// MustParseMarkdown parses markdown into a document node.
func MustParseMarkdown(t *testing.T, md goldmark.Markdown, ctx parser.Context, src string) ast.Node {
	t.Helper()
	reader := text.NewReader([]byte(src))
	doc := md.Parser().Parse(reader, parser.WithContext(ctx))
	if errs := mdctx.PopErrors(ctx); len(errs) > 0 {
		t.Fatalf("errors while parsing: %v", errs)
	}
	return doc
}

// AssertNoRenderDiff asserts the markdown instance renders src into the wanted
// string.
func AssertNoRenderDiff(t *testing.T, doc ast.Node, md goldmark.Markdown, src, want string, opts ...cmp.Option) {
	t.Helper()
	bufW := &bytes.Buffer{}
	if err := md.Renderer().Render(bufW, []byte(src), doc); err != nil {
		t.Fatal(err)
	}

	if diff, err := htmls.Diff(strings.NewReader(want), bufW, opts...); err != nil {
		t.Fatal(err)
	} else if diff != "" {
		t.Errorf("Render mismatch (-want +got):\n%s", diff)
	}
}
