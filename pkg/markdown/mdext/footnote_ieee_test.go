package mdext

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/htmls/tags"
	"github.com/jschaf/b2/pkg/markdown/mdtest"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/jschaf/bibtex"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

func newCiteIEEE(key bibtex.CiteKey, order string) string {
	return newCiteIEEECount(key, order, 0)
}

const testPath = "/abs-path/"

var (
	previewRegex    = regexp.MustCompile(` data-preview-snippet=".*?"`)
	styleRegex      = regexp.MustCompile(` style=".*?"`)
	beforeRefsRegex = regexp.MustCompile(`(?s).*?(<div.*)`)
)

var removePreviewOpt = cmp.Transformer("removePreviewOpt", func(s string) string {
	s1 := previewRegex.ReplaceAllString(s, "")
	return styleRegex.ReplaceAllString(s1, "")
})

var removeAllButReferences = cmp.Transformer("removeAllButReferences", func(s string) string {
	return beforeRefsRegex.ReplaceAllString(s, "$1")
})

func newCiteIEEECount(key bibtex.CiteKey, order string, count int) string {
	id := "footnote-link-" + key
	if count > 0 {
		id += "-" + strconv.Itoa(count)
	}
	attrs := []string{
		`data-link-type=citation`,
		`class="preview-target footnote-link"`,
		fmt.Sprintf(`href="%s#cite_ref_%s"`, testPath, key),
		`role=doc-noteref`,
		`id=` + id,
	}
	return tags.AAttrs(strings.Join(attrs, " "), tags.Cite(order))
}

func newCiteIEEEAside(key bibtex.CiteKey, count int, t ...string) string {
	id := "footnote-body-" + key
	if count > 1 {
		id += "-" + strconv.Itoa(count-1)
	}
	attrs := []string{
		`class="footnote-body-cite footnote-body"`,
		`id=` + id,
		`role=doc-endnote`,
		`style="margin-top: -18px"`,
	}
	return tags.AsideAttrs(strings.Join(attrs, " "),
		t...,
	)
}

func newInlineCite(order string) string {
	return tags.CiteAttrs("class=cite-inline", order)
}

func joinElems(t ...string) string {
	return strings.Join(t, "\n")
}

func TestNewFootnoteExt_IEEE(t *testing.T) {
	style := cite.IEEE
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"keeps surrounding text",
			"alpha [^@bib_foo] bravo",
			joinElems(
				tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo"),
				newCiteIEEEAside("bib_foo", 1, tags.P(newInlineCite("[1]"), newBibFooCite())),
			),
		},
		{
			"numbers different citations",
			"alpha [^@bib_foo] bravo [^@bib_bar]",
			joinElems(
				tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo ", newCiteIEEE("bib_bar", "[2]")),
				newCiteIEEEAside("bib_foo", 1, tags.P(newInlineCite("[1]"), newBibFooCite())),
				newCiteIEEEAside("bib_bar", 1, tags.P(newInlineCite("[2]"), newBibBarCite())),
			),
		},
		{
			"re-use citation numbers",
			"alpha [^@bib_foo] bravo [^@bib_bar] charlie [^@bib_foo] delta [^@bib_bar]",
			joinElems(
				tags.P(
					"alpha ", newCiteIEEE("bib_foo", "[1]"),
					" bravo ", newCiteIEEE("bib_bar", "[2]"),
					" charlie ", newCiteIEEECount("bib_foo", "[1]", 1),
					" delta ", newCiteIEEECount("bib_bar", "[2]", 1),
				),
				newCiteIEEEAside("bib_foo", 1, tags.P(newInlineCite("[1]"), newBibFooCite())),
				newCiteIEEEAside("bib_bar", 1, tags.P(newInlineCite("[2]"), newBibBarCite())),
				newCiteIEEEAside("bib_foo", 2, tags.P(newInlineCite("[1]"), newBibFooCite())),
				newCiteIEEEAside("bib_bar", 2, tags.P(newInlineCite("[2]"), newBibBarCite())),
			),
		},
		{
			"order numbering for mix of footnote and citation",
			texts.Dedent(`
        alpha [^@bib_foo] [^side:foo] 

        ::: footnote side:foo
        body-text
        :::

        bravo [^@bib_bar]
			`),
			joinElems(
				tags.P(
					"alpha ",
					newCiteIEEE("bib_foo", "[1]"),
					`<a href="#footnote-body-side:foo" class="footnote-link" role="doc-noteref" id="footnote-link-side:foo">`,
					`<cite>[2]</cite>`,
					`</a>`,
				),
				texts.Dedent(`
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote" style="margin-top: -18px">
        <p><cite class=cite-inline>[2]</cite> body-text</p>
        </aside>
			`),
				newCiteIEEEAside("bib_foo", 1, tags.P(newInlineCite("[1]"), newBibFooCite())),
				tags.P(
					"bravo ",
					newCiteIEEE("bib_bar", "[3]"),
				),
				newCiteIEEEAside("bib_bar", 1, tags.P(newInlineCite("[3]"), newBibBarCite())),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewFootnoteExt(style, NewCitationNopAttacher()),
				NewColonBlockExt(), // footnote bodies are colon blocks
				NewCustomExt(),     // cite tags are implemented via custom
			)
			SetTOMLMeta(ctx, PostMeta{
				BibPaths: []string{"./testdata/citation_test.bib"},
				Path:     testPath,
			})
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want, removePreviewOpt)
		})
	}
}

func TestNewCitationExt_IEEE_renderMultiplePosts(t *testing.T) {
	style := cite.IEEE
	md1 := "alpha [^@bib_foo] bravo"
	want1 := joinElems(
		tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo"),
		newCiteIEEEAside("bib_foo", 1, tags.P(newInlineCite("[1]"), newBibFooCite())),
	)
	md2 := "alpha [^@bib_foo] bravo [^@bib_bar]"
	want2 := joinElems(
		tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo ", newCiteIEEE("bib_bar", "[2]")),
		newCiteIEEEAside("bib_foo", 1, tags.P(newInlineCite("[1]"), newBibFooCite())),
		newCiteIEEEAside("bib_bar", 1, tags.P(newInlineCite("[2]"), newBibBarCite())),
	)

	mdTester, _ := mdtest.NewTester(t,
		NewFootnoteExt(style, NewCitationNopAttacher()),
		NewColonBlockExt(), // footnote bodies are colon blocks
		NewCustomExt(),     // cite tags are implemented via custom
	)

	t.Run("first run", func(t *testing.T) {
		ctx1 := parser.NewContext()
		SetTOMLMeta(ctx1, PostMeta{
			BibPaths: []string{"./testdata/citation_test.bib"},
			Path:     testPath,
		})
		doc1 := mdtest.MustParseMarkdown(t, mdTester, ctx1, md1)
		mdtest.AssertNoRenderDiff(t, doc1, mdTester, md1, want1, removePreviewOpt)
	})

	t.Run("second run", func(t *testing.T) {
		ctx2 := parser.NewContext()
		SetTOMLMeta(ctx2, PostMeta{
			BibPaths: []string{"./testdata/citation_test.bib"},
			Path:     testPath,
		})
		doc2 := mdtest.MustParseMarkdown(t, mdTester, ctx2, md2)
		mdtest.AssertNoRenderDiff(t, doc2, mdTester, md2, want2, removePreviewOpt)
	})
}

type citeDocAttacher struct{}

func (c citeDocAttacher) Attach(doc *ast.Document, refs *CitationReferences) error {
	doc.AppendChild(doc, refs)
	return nil
}

// newCiteRefsIEEE creates the div containing references.
func newCiteRefsIEEE(ts ...string) string {
	return tags.DivAttrs("class=cite-references",
		tags.H2("References"),
		strings.Join(ts, ""))
}

func newCiteRefIEEE(key bibtex.CiteKey, count int, order string, content ...string) string {
	c := &Citation{Key: key}
	orderN, err := strconv.Atoi(order[1 : len(order)-1])
	if err != nil {
		panic(err)
	}
	cr := &CitationRef{
		Citation: c,
		Order:    orderN,
		Count:    count,
	}
	divAttrs := fmt.Sprintf(`id=%s class=cite-reference`, c.ReferenceID())
	citeAttrs := `class=preview-target data-link-type=cite-reference-num data-cite-ids="` +
		strings.Join(allCiteIDs(cr), " ") + `"`
	return tags.DivAttrs(divAttrs,
		tags.CiteAttrs(citeAttrs, order),
		strings.Join(content, ""))
}

func newJournal(ts ...string) string {
	return tags.EmAttrs("class=cite-journal", ts...)
}

func newBibFooCite() string {
	return strings.Join([]string{
		"F. Q. Blogs, J. P. Doe and A. Idiot,",
		`"Turtles in the time continuum," in`,
		newJournal("Turtles in the Applied Sciences"),
		", Vol. 3, 2016.",
	}, " ")
}

func newBibBarCite() string {
	return strings.Join([]string{
		"E. Ortiz, J. Breads and C. Clarisse,",
		`"Turtles in the time continuum," in`,
		newJournal("Nature"),
		", Vol. 3, 2019.",
	}, " ")
}

func TestNewCitationExt_IEEE_References(t *testing.T) {
	style := cite.IEEE
	tests := []struct {
		name     string
		src      string
		wantRefs string
	}{
		{
			"2 references",
			"alpha [^@bib_foo] bravo [^@bib_bar] charlie [^@bib_foo] delta [^@bib_bar]",
			newCiteRefsIEEE(
				newCiteRefIEEE("bib_foo", 2, "[1]", newBibFooCite()),
				newCiteRefIEEE("bib_bar", 2, "[2]", newBibBarCite()),
			),
		},
		{
			"concurrent programming in java",
			"alpha [^@lea2000concurrent]",
			newCiteRefsIEEE(
				newCiteRefIEEE("lea2000concurrent", 1, "[1]",
					`D. Lea, "Concurrent Programming in Java: Design Principles and Patterns," 2000.`,
				),
			),
		},
		{
			"corbett2012spanner",
			"[^@corbett2012spanner]",
			newCiteRefsIEEE(
				newCiteRefIEEE("corbett2012spanner", 1, "[1]",
					`J. C. Corbett, <em>et al.</em>, "Spanner: Google's Globally-Distributed Database," 2012.`,
				),
			),
		},
		{
			"corbett2013spanner",
			"[^@corbett2013spanner]",
			newCiteRefsIEEE(
				newCiteRefIEEE("corbett2013spanner", 1, "[1]",
					`J. C. Corbett, <em>et al.</em>, "Spanner: Google's Globally-Distributed Database,"`,
					" in "+newJournal("ACM Trans. Comput. Syst.")+",",
					" Vol. 31,",
					" 2013,",
					" doi: ",
					tags.AAttrs(`href="https://doi.org/10.1145/2491245"`, "10.1145/2491245"),
					".",
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t,
				NewFootnoteExt(style, citeDocAttacher{}),
				NewColonBlockExt(), // footnote bodies are colon blocks
				NewCustomExt(),     // cite tags are implemented via custom
			)
			SetTOMLMeta(ctx, PostMeta{
				BibPaths: []string{"./testdata/citation_test.bib"},
				Path:     testPath,
			})
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.wantRefs, removeAllButReferences)
		})
	}
}
