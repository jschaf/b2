package mdext

import (
	"bytes"
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
        <aside class="footnote-body" id="footnote-body-side:foo" role="doc-endnote">
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
		newJournal("Turtles Appl. Sciences"),
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
			"corbett2012spanner",
			"[^@corbett2012spanner]",
			newCiteRefsIEEE(
				newCiteRefIEEE("corbett2012spanner", 1, "[1]",
					`J. C. Corbett, <em>et al.</em>, "Spanner: Google's Globally-Distributed Database," in <em class="cite-conference">OSDI</em>, 2012.`,
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

// testBufWriter is a simple implementation of util.BufWriter.
type testBufWriter struct {
	*bytes.Buffer
}

func newTestBufWriter() *testBufWriter {
	return &testBufWriter{
		Buffer: &bytes.Buffer{},
	}
}

func (tw *testBufWriter) Available() int {
	return tw.Cap() - tw.Len()
}

func (tw *testBufWriter) Buffered() int {
	return tw.Len()
}

func (tw *testBufWriter) Flush() error {
	return nil
}

func ieeePageRange(lo, hi int) string {
	return "pp. " + strconv.Itoa(lo) + texts.EnDash + strconv.Itoa(hi)
}

func ieeeDOI(doi string) string {
	return fmt.Sprintf(`doi: <a href="https://doi.org/%s">%s</a>`, doi, doi)
}

func ieeeJournal(journal string) string {
	return fmt.Sprintf(`in <em class=cite-journal>%s</em>`, journal)
}

func ieeeBook(book string) string {
	return fmt.Sprintf(`<em class=cite-book>%s</em>`, book)
}

func TestNewFootnoteExt_renderCiteRefContent(t *testing.T) {
	tests := []struct {
		bibEntry string
		want     string
	}{
		{
			texts.Dedent(`
				@inproceedings{canonne2020learning,
				  title={Learning from satisfying assignments under continuous distributions},
				  author={Canonne, Clement L and De, Anindya and Servedio, Rocco A},
				  booktitle={Proceedings of the Fourteenth Annual ACM-SIAM Symposium on Discrete Algorithms},
				  pages={82--101},
				  year={2020},
				  organization={SIAM}
			  }
     	`),
			texts.Join(
				`C. L. Canonne, A. De and R. A. Servedio, `,
				`"Learning from satisfying assignments under continuous distributions," `,
				`in <em class=cite-conference>Proc. 14th Annu. ACM-SIAM Symp. Discrete Algorithms</em>, `,
				`2020, `,
				ieeePageRange(82, 101),
				".",
			),
		},
		{
			texts.Dedent(`
        @article{badrin2017segnet,
        	title        = {SegNet: A Deep Convolutional Encoder-Decoder Architecture for Image Segmentation},
        	author       = {V. {Badrinarayanan} and A. {Kendall} and R. {Cipolla}},
        	year         = 2017,
        	journal      = {IEEE Transactions on Pattern Analysis and Machine Intelligence},
        	volume       = 39,
        	number       = 12,
        	pages        = {2481--2495},
        	doi          = {10.1109/TPAMI.2016.2644615}
        }
			`),
			texts.JoinSpace(
				`V. Badrinarayanan, A. Kendall and R. Cipolla,`,
				`"SegNet: A Deep Convolutional Encoder-Decoder Architecture for Image Segmentation,"`,
				ieeeJournal(`IEEE Trans. Pattern Analysis Mach. Intell.`)+`,`,
				`Vol. 39,`,
				`no. 12,`,
				`2017,`,
				ieeePageRange(2481, 2495)+`,`,
				ieeeDOI(`10.1109/TPAMI.2016.2644615`)+`.`,
			),
		},
		{
			texts.Dedent(`
        @book{raj1991art,
        	title        = {The Art of Computer Systems Performance Analysis},
			    subtitle     = {Techniques for Experimental Design, Measurement, Simulation, and Modeling},
        	author       = {Jain, Raj},
        	year         = 1991,
        	publisher    = {Wiley},
        	series       = {Wiley professional computing},
        	isbn         = {978-0-471-50336-1},
        }
			`),
			texts.JoinSpace(
				`R. Jain,`,
				ieeeBook(`The Art of Computer Systems Performance Analysis`)+`,`,
				`Wiley,`,
				`1991.`,
			),
		},
	}
	for _, tt := range tests {
		t.Run(texts.FirstLine(strings.TrimSpace(tt.bibEntry)), func(t *testing.T) {
			biber := cite.Biber
			bibAST, err := biber.Parse(strings.NewReader(tt.bibEntry))
			if err != nil {
				t.Error(err)
			}
			entries, err := biber.Resolve(bibAST)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatalf("expected exactly 1 bibtex entry, had %d", len(entries))
			}
			entry := entries[0]
			c := NewCitation()
			c.Key = entry.Key
			c.Bibtex = entry
			b := newTestBufWriter()
			renderCiteRefContent(b, c)
			if diff := cmp.Diff(tt.want, b.String()); diff != "" {
				t.Errorf("renderCiteRefContent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
