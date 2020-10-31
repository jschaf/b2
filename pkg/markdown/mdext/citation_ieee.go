package mdext

import "C"
import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/bibtex"
	"github.com/jschaf/b2/pkg/markdown/asts"
	"github.com/jschaf/b2/pkg/markdown/attrs"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// AST Transformer to update the cite ID and add HTML tag attributes.
type citationIEEEFormatTransformer struct{}

func (c citationIEEEFormatTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	// The number of times a bibtex cite key has been used thus far. Useful for
	// generating unique IDs for the citation.
	citeCounts := make(map[bibtex.CiteKey]int)

	err := asts.WalkKind(KindCitation, doc, func(n ast.Node) (ast.WalkStatus, error) {
		c := n.(*Citation)
		c.SetAttribute([]byte("href"), c.AbsPath+"/#"+c.ReferenceID())
		// Create the HTML preview on the <a> tag.
		b := &bytes.Buffer{}
		citeHTML := bufio.NewWriter(b)
		citeHTML.WriteString("<p>")
		renderCiteRefContent(citeHTML, c)
		citeHTML.WriteString("</p>")
		citeHTML.Flush()
		attrs.AddClass(c, "preview-target")
		c.SetAttribute([]byte("data-preview-snippet"), b.Bytes())
		c.SetAttribute([]byte("data-link-type"), LinkCitation)
		c.ID = c.CiteID(citeCounts[c.Key])
		citeCounts[c.Key] += 1
		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		mdctx.PushError(pc, fmt.Errorf("walk IEEE cite format transformer: %w", err))
	}
}

// citationRendererIEEE renders an IEEE citation.
type citationRendererIEEE struct {
	includeRefs bool
}

func (cr *citationRendererIEEE) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCitation, cr.renderCitation)
	if cr.includeRefs {
		reg.Register(KindCitationReferences, cr.renderReferenceList)
	} else {
		reg.Register(KindCitationReferences, asts.NopRender)
	}
}

func (cr *citationRendererIEEE) renderCitation(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	c := n.(*Citation)
	w.WriteString(`<a`)
	attrs.RenderAll(w, c)
	w.WriteByte('>')
	w.WriteString(`<cite id=`)
	w.WriteString(c.ID)
	w.WriteByte('>')
	w.WriteString("[" + strconv.Itoa(c.Order) + "]")
	w.WriteString("</cite>")
	w.WriteString("</a>")
	// Citations generate content solely from the citation, not children.
	return ast.WalkSkipChildren, nil
}

func (cr *citationRendererIEEE) renderReferenceList(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	refs := n.(*CitationReferences)
	if len(refs.Citations) == 0 {
		return ast.WalkSkipChildren, nil
	}

	hasRef := make(map[bibtex.CiteKey]struct{})
	counts := getCiteCounts(refs)

	_, _ = w.WriteString(`<div class=cite-references>`)
	_, _ = w.WriteString(`<h2>References</h2>`)
	for _, c := range refs.Citations {
		if _, ok := hasRef[c.Key]; ok {
			continue
		}
		hasRef[c.Key] = struct{}{}
		cr.renderCiteRef(w, c, counts)
	}
	_, _ = w.WriteString(`</div>`)

	return ast.WalkContinue, nil
}

// allCiteIDs returns a slice of strings where each string is an HTML ID of a
// citation.
func allCiteIDs(c *Citation, count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = c.CiteID(i)
	}
	return ids
}

// getCiteCounts returns a map showing the number of times a citation appeared
// in the document. Useful for backlinks from references to the citation.
func getCiteCounts(refs *CitationReferences) map[bibtex.CiteKey]int {
	cnts := make(map[bibtex.CiteKey]int, len(refs.Citations))
	for _, c := range refs.Citations {
		cnts[c.Key] += 1
	}
	return cnts
}

func (cr *citationRendererIEEE) renderCiteRef(w util.BufWriter, c *Citation, counts map[bibtex.CiteKey]int) {
	cnt := counts[c.Key]
	citeIDs := allCiteIDs(c, cnt)
	w.WriteString(`<div id="`)
	w.WriteString(c.ReferenceID())
	w.WriteString(`" class=cite-reference>`)
	w.WriteString(`<cite class=preview-target data-link-type=cite-reference-num data-cite-ids="`)
	for i, c := range citeIDs {
		if i > 0 {
			w.WriteByte(' ')
		}
		w.WriteString(c)
	}
	w.WriteString(`">`)
	w.WriteString("[" + strconv.Itoa(c.Order) + "]")
	w.WriteString(`</cite> `)

	renderCiteRefContent(w, c)
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
