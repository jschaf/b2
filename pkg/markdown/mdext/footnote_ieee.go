package mdext

import "C"
import (
	"github.com/jschaf/b2/pkg/markdown/asts"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/jschaf/bibtex"
	"github.com/yuin/goldmark/renderer"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// footnoteIEEERenderer renders an IEEE citation.
type footnoteIEEERenderer struct {
	includeRefs bool
}

func (fr *footnoteIEEERenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCitation, fr.renderCitation)
	if fr.includeRefs {
		reg.Register(KindCitationReferences, fr.renderReferenceList)
	} else {
		reg.Register(KindCitationReferences, asts.NopRender)
	}
}

func (fr *footnoteIEEERenderer) renderCitation(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	c := n.(*Citation)
	renderCiteRefContent(w, c)
	// Citations generate content solely from the citation, not children.
	return ast.WalkSkipChildren, nil
}

func (fr *footnoteIEEERenderer) renderReferenceList(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	refs := n.(*CitationReferences)
	if len(refs.Refs) == 0 {
		return ast.WalkSkipChildren, nil
	}

	_, _ = w.WriteString(`<div class=cite-references>`)
	_, _ = w.WriteString(`<h2>References</h2>`)
	for _, c := range refs.Refs {
		fr.renderCiteRef(w, c)
	}
	_, _ = w.WriteString(`</div>`)

	return ast.WalkContinue, nil
}

// allCiteIDs returns a slice of strings where each string is an HTML ID of a
// citation.
func allCiteIDs(cr *CitationRef) []string {
	ids := make([]string, cr.Count)
	for i := range ids {
		ids[i] = cr.Citation.CiteID(i)
	}
	return ids
}

func (fr *footnoteIEEERenderer) renderCiteRef(w util.BufWriter, cr *CitationRef) {
	citeIDs := allCiteIDs(cr)
	w.WriteString(`<div id="`)
	w.WriteString(cr.Citation.ReferenceID())
	w.WriteString(`" class=cite-reference>`)
	w.WriteString(`<cite class=preview-target data-link-type=cite-reference-num data-cite-ids="`)
	for i, c := range citeIDs {
		if i > 0 {
			w.WriteByte(' ')
		}
		w.WriteString(c)
	}
	w.WriteString(`">`)
	w.WriteString("[" + strconv.Itoa(cr.Order) + "]")
	w.WriteString(`</cite> `)

	renderCiteRefContent(w, cr.Citation)
	w.WriteString(`</div>`)
}

func renderCiteRefContent(w util.BufWriter, c *Citation) {
	renderAuthors(w, c)
	renderTitle(w, c)
	hasInfoAfterTitle := false
	writeTitleSep := func() {
		if hasInfoAfterTitle {
			w.WriteRune(',')
		}
		w.WriteRune(' ')
	}

	if jrn := trimBraces(c.Bibtex.Tags["journal"]); jrn != "" {
		writeTitleSep()
		renderJournal(w, jrn)
		hasInfoAfterTitle = true
	}

	if vol := trimBraces(c.Bibtex.Tags["volume"]); vol != "" {
		writeTitleSep()
		w.WriteString("Vol. ")
		w.WriteString(vol)
		hasInfoAfterTitle = true
	}

	if year := trimBraces(c.Bibtex.Tags["year"]); year != "" {
		writeTitleSep()
		w.WriteString(year)
		hasInfoAfterTitle = true
	}

	if doi := trimBraces(c.Bibtex.Tags["doi"]); doi != "" {
		writeTitleSep()
		renderDOI(w, c)
		hasInfoAfterTitle = true
	}
	w.WriteString(".")
}

func renderAuthors(w util.BufWriter, c *Citation) {
	authors := c.Bibtex.Author
	// IEEE Ref, Sec II: If there are more than six names listed, use the primary
	// author's name followed by et al.
	if len(authors) > 6 {
		renderAuthor(w, authors[0])
		w.WriteString(", <em>et al.</em>")
		return
	}

	for i, author := range authors {
		renderAuthor(w, author)
		if i < len(authors)-2 {
			w.WriteString(", ")
		} else if i == len(authors)-2 {
			if authors[len(authors)-1].IsOthers() {
				w.WriteString(" <em>et al.</em>")
				break
			} else {
				w.WriteString(" and ")
			}
		}
	}
}

func renderAuthor(w util.BufWriter, author bibtex.Author) {
	// IEEE Ref, Sec II: In all references, the given name of the author or editor
	// is abbreviated to the initial only and precedes the last name. Use commas
	// around Jr., Sr., and III in names.
	sp := strings.Split(author.First, " ")
	for _, s := range sp {
		if r, _ := utf8.DecodeRuneInString(s); r != utf8.RuneError {
			w.WriteRune(r)
			w.WriteString(". ")
		}
	}
	w.WriteString(author.Last)
}

func renderTitle(w util.BufWriter, c *Citation) {
	title := trimBraces(c.Bibtex.Tags["title"])
	w.WriteString(`, "`)
	w.WriteString(title)
	w.WriteString(`,"`)
}

func renderJournal(w util.BufWriter, journal string) {
	w.WriteString("in <em class=cite-journal>")
	w.WriteString(journal)
	w.WriteString("</em>")
}

func renderDOI(w util.BufWriter, c *Citation) {
	w.WriteString("doi: ")
	doi := trimBraces(c.Bibtex.Tags["doi"])
	w.WriteString(`<a href="https://doi.org/`)
	w.WriteString(doi)
	w.WriteString(`">`)
	w.WriteString(doi)
	w.WriteString(`</a>`)
}

func trimBraces(s string) string {
	b := texts.ReadOnlyBytes(s)
	lo, hi := 0, len(b)
	for ; lo < len(b); lo++ {
		if b[lo] != '{' {
			break
		}
	}
	for ; hi > 0; hi-- {
		if b[hi-1] != '}' {
			break
		}
	}
	return texts.ReadonlyString(b[lo:hi])
}
