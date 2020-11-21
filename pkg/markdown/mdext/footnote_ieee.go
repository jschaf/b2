//
package mdext

import "C"
import (
	"fmt"
	"github.com/jschaf/b2/pkg/markdown/asts"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/jschaf/bibtex"
	bibast "github.com/jschaf/bibtex/ast"
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

// renderCiteRefContent is the main formatter for an IEEE citation.
// Follows https://ieeeauthorcenter.ieee.org/wp-content/uploads/IEEE-Reference-Guide.pdf.
func renderCiteRefContent(w util.BufWriter, c *Citation) {
	if au := c.Bibtex.Tags[bibtex.FieldAuthor]; au != nil {
		renderAuthors(w, au)
	}
	renderTitle(w, c)
	hasInfoAfterTitle := false
	writeSep := func() {
		if hasInfoAfterTitle {
			w.WriteRune(',')
		}
		w.WriteRune(' ')
		hasInfoAfterTitle = true
	}

	switch c.Bibtex.Type {
	case bibtex.EntryArticle:
		if jrn := c.Bibtex.Tags[bibtex.FieldJournal]; jrn != nil {
			writeSep()
			renderJournal(w, jrn)
		}
	case bibtex.EntryInProceedings:
		if t := c.Bibtex.Tags[bibtex.FieldBookTitle]; t != nil {
			writeSep()
			renderConference(w, t)
		}
	}

	if vol := c.Bibtex.Tags[bibtex.FieldVolume]; vol != nil {
		writeSep()
		w.WriteString("Vol. ")
		w.WriteString(assertSimpleText(vol))
	}

	if num := c.Bibtex.Tags[bibtex.FieldNumber]; num != nil {
		writeSep()
		w.WriteString("no. ")
		w.WriteString(assertSimpleText(num))
	}

	if c.Bibtex.Type == bibtex.EntryBook {
		if pub := c.Bibtex.Tags[bibtex.FieldPublisher]; pub != nil {
			writeSep()
			w.WriteString(assertSimpleText(pub))
		}
	}

	if year := c.Bibtex.Tags[bibtex.FieldYear]; year != nil {
		writeSep()
		w.WriteString(assertSimpleText(year))
	}

	if p := c.Bibtex.Tags[bibtex.FieldPages]; p != nil {
		writeSep()
		w.WriteString("pp. ")
		w.WriteString(strings.Replace(assertSimpleText(p), "--", texts.EnDash, 1))
	}

	if doi := c.Bibtex.Tags["doi"]; doi != nil {
		writeSep()
		renderDOI(w, doi)
	}
	w.WriteString(".")
}

func renderAuthors(w util.BufWriter, x bibast.Expr) {
	authors, ok := x.(bibast.Authors)
	if !ok {
		panic(fmt.Sprintf("render authors want bibast.Authors; got %T", x))
	}
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

func renderAuthor(w util.BufWriter, author *bibast.Author) {
	// IEEE Ref, Sec II: In all references, the given name of the author or editor
	// is abbreviated to the initial only and precedes the last name. Use commas
	// around Jr., Sr., and III in names.
	sp := strings.Split(assertSimpleText(author.First), " ")
	for _, s := range sp {
		if r, _ := utf8.DecodeRuneInString(s); r != utf8.RuneError {
			w.WriteRune(r)
			w.WriteString(". ")
		}
	}
	w.WriteString(assertSimpleText(author.Last))
}

func renderTitle(w util.BufWriter, c *Citation) {
	t := c.Bibtex.Tags[bibtex.FieldTitle]
	if t == nil {
		return
	}
	title := assertSimpleText(t)
	hasURL := c.Bibtex.Tags["url"] != nil
	openURL := func() {
		if hasURL {
			w.WriteString(`<a href="`)
			url := assertSimpleText(c.Bibtex.Tags["url"])
			w.Write(util.EscapeHTML([]byte(url)))
			w.WriteString(`">`)
		}
	}
	closeURL := func() {
		if hasURL {
			w.WriteString("</a>")
		}
	}

	switch c.Bibtex.Type {
	case bibtex.EntryBook:
		w.WriteString(`, <em class=cite-book>`)
		openURL()
		w.WriteString(title)
		closeURL()
		w.WriteString(`</em>,`)
	default:
		w.WriteString(`, "`)
		openURL()
		w.WriteString(title)
		closeURL()
		w.WriteString(`,"`)
	}
}
func assertSimpleText(x bibast.Expr) string {
	if x == nil {
		panic("expected simple text for bibast.Expr but was nil")
	}
	switch txt := x.(type) {
	case *bibast.Text:
		return txt.Value
	case *bibast.Number:
		return txt.Value
	default:
		panic(fmt.Sprintf("unsupported bibast type for bibast.Expr; got %T", x))
	}
}

func renderJournal(w util.BufWriter, journal bibast.Expr) {
	w.WriteString("in <em class=cite-journal>")
	_, _ = ieeeAbbrevReplacer.WriteString(w, assertSimpleText(journal))
	w.WriteString("</em>")
}

// renderConference formats a conference name in IEEE style.
//
// IEEE reference guide, section B includes the following guidance:
//
// 1. Use the standard abbreviations below for all words in the conference.
// 2. Write out all the remaining words, but omit most articles and prepositions
//    like "of the" and "on." That is,
//    "Proceedings of the 1996 Robotics and Automation Conference"
//    becomes
//    "Proc. 1996 Robotics and Automation Conf."
// 3. All published conference or proceedings papers have page numbers.
func renderConference(w util.BufWriter, c bibast.Expr) {
	w.WriteString("in <em class=cite-conference>")
	_, _ = ieeeAbbrevReplacer.WriteString(w, assertSimpleText(c))
	w.WriteString("</em>")
}

func renderDOI(w util.BufWriter, doi bibast.Expr) {
	doiTxt := assertSimpleText(doi)
	w.WriteString("doi: ")
	w.WriteString(`<a href="https://doi.org/`)
	w.WriteString(doiTxt)
	w.WriteString(`">`)
	w.WriteString(doiTxt)
	w.WriteString(`</a>`)
}

var ieeeAbbrevReplacer = strings.NewReplacer(
	// Common abbreviations.
	"Annals", "Ann.",
	"Annual", "Annu.",
	"Applied", "Appl.",
	"Colloquium", "Colloq.",
	"Communications", "Commun.",
	"Conference", "Conf.",
	"Congress", "Congr.",
	"Convention", "Conv.",
	"Digest", "Dig.",
	"Exposition", "Expo.",
	"Intelligence", "Intell.",
	"International", "Int.",
	"Journal", "J.",
	"Machine", "Mach.",
	"National", "Nat.",
	"Proceedings", "Proc.",
	"Record", "Rec.",
	"Society", "Soc.",
	"Systems", "Syst.",
	"Symposium", "Symp.",
	"Technical", "Tech.",
	"Transactions", "Trans.",
	// Replace numbers with ordinals. The general case requires a custom replacer
	// so hard-code common numbers instead.
	"First", "1st",
	"Second", "2nd",
	"Third", "3rd",
	"Fourth", "4th",
	"Fifth", "5th",
	"Sixth", "6th",
	"Seventh", "7th",
	"Eighth", "8th",
	"Ninth", "9th",
	"Tenth", "10th",
	"Eleventh", "11th",
	"Twelfth", "12th",
	"Thirteenth", "13th",
	"Fourteenth", "14th",
	"Fifteenth", "15th",
	"Sixteenth", "16th",
	"Seventeenth", "17th",
	"Eighteenth", "18th",
	"Nineteenth", "19th",
	// To replace articles, we need to anchor with spaces. This isn't a perfect
	// way to replace all articles but it's good enough. The best method is to
	// write our own replacer. Multiple runs of articles must be replaced
	// separately.
	// Two word articles. Must come before single word articles.
	" in the ", " ",
	" of the ", " ",
	" of a ", " ",
	// Single word articles
	" a ", " ",
	" and ", " ",
	" by ", " ",
	" from ", " ",
	" in ", " ",
	" of ", " ",
	" on ", " ",
	" the ", " ",
	" to ", " ",
	" with ", " ",
)
