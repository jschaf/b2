package mdext

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/bibtex"
	"github.com/jschaf/b2/pkg/cite"
	"github.com/jschaf/b2/pkg/htmls/tags"
	"github.com/yuin/goldmark/ast"
)

func newCiteIEEE(key bibtex.CiteKey, order string) string {
	return newCiteIEEECount(key, order, 0)
}

func newCiteIEEECount(key bibtex.CiteKey, order string, count int) string {
	id := "cite_" + key
	if count > 0 {
		id += "_" + strconv.Itoa(count)
	}
	attrs := fmt.Sprintf(`id=%s`, id)
	aAttrs := fmt.Sprintf(
		`href="%s" class=preview-target data-link-type=citation`,
		"#cite_ref_"+key)
	return tags.AAttrs(aAttrs, tags.CiteAttrs(attrs, order))
}

func TestNewCitationExt_IEEE(t *testing.T) {
	style := cite.IEEE
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"ignores prefix and suffix",
			"[**qux**, @bib_foo *bar*]",
			tags.P(newCiteIEEE("bib_foo", "[1]")),
		},
		{
			"keeps surrounding text",
			"alpha [@bib_foo] bravo",
			tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo"),
		},
		{
			"numbers different citations",
			"alpha [@bib_foo] bravo [@bib_bar]",
			tags.P("alpha ", newCiteIEEE("bib_foo", "[1]"), " bravo ", newCiteIEEE("bib_bar", "[2]")),
		},
		{
			"re-use citation numbers",
			"alpha [@bib_foo] bravo [@bib_bar] charlie [@bib_foo] delta [@bib_bar]",
			tags.P(
				"alpha ", newCiteIEEE("bib_foo", "[1]"),
				" bravo ", newCiteIEEE("bib_bar", "[2]"),
				" charlie ", newCiteIEEECount("bib_foo", "[1]", 1),
				" delta ", newCiteIEEECount("bib_bar", "[2]", 1),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := newMdTester(t, NewCitationExt(style, NewCitationNopAttacher()))
			SetTOMLMeta(ctx, PostMeta{
				BibPaths: []string{"./testdata/citation_test.bib"},
			})
			doc := mustParseMarkdown(t, md, ctx, tt.src)
			assertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
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
	divAttrs := fmt.Sprintf(`id=%s class=cite-reference`, c.ReferenceID())
	citeAttrs := `class=preview-target data-link-type=cite-reference-num data-cite-ids="` +
		strings.Join(allCiteIDs(c, count), " ") + `"`
	return tags.DivAttrs(divAttrs,
		tags.CiteAttrs(citeAttrs, order),
		strings.Join(content, ""))
}

func newJournal(ts ...string) string {
	return tags.EmAttrs("class=cite-journal", ts...)
}

func TestNewCitationExt_IEEE_References(t *testing.T) {
	style := cite.IEEE
	tests := []struct {
		name     string
		src      string
		wantBody string
		wantRefs string
	}{
		{
			"2 references",
			"alpha [@bib_foo] bravo [@bib_bar] charlie [@bib_foo] delta [@bib_bar]",
			tags.P(
				"alpha ", newCiteIEEE("bib_foo", "[1]"),
				" bravo ", newCiteIEEE("bib_bar", "[2]"),
				" charlie ", newCiteIEEECount("bib_foo", "[1]", 1),
				" delta ", newCiteIEEECount("bib_bar", "[2]", 1),
			),
			newCiteRefsIEEE(
				newCiteRefIEEE("bib_foo", 2, "[1]",
					"F. Q. Bloggs, J. P. Doe and A. Idiot, ",
					`"Turtles in the time continum," in`,
					newJournal("Turtles in the Applied Sciences"),
					", Vol. 3, 2016.",
				),
				newCiteRefIEEE("bib_bar", 2, "[2]",
					"E. Orti, J. Bredas and C. Clarisse, ",
					`"Turtles in the time continum," in`,
					newJournal("Nature"),
					", Vol. 3, 2019.",
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := newMdTester(t, NewCitationExt(style, citeDocAttacher{}))
			SetTOMLMeta(ctx, PostMeta{
				BibPaths: []string{"./testdata/citation_test.bib"},
			})
			doc := mustParseMarkdown(t, md, ctx, tt.src)
			assertNoRenderDiff(t, doc, md, tt.src, tt.wantBody+"\n"+tt.wantRefs)
		})
	}
}
