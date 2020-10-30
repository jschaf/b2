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

type citationIEEEFormatTransformer struct{}

func (c citationIEEEFormatTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	// Next number to use as a citation reference, like [1]. Starts at 1.
	nextNum := 1
	// Mapping from the abs path to the bibtex cite key to the order the first
	// instance of a cite key appeared in the markdown document.
	citeNums := make(map[bibtex.CiteKey]int)
	// The number of times a bibtex cite key has been used thus far. Useful for
	// generating unique IDs for the citation.
	citeCounts := make(map[bibtex.CiteKey]int)

	err := asts.WalkKind(KindCitation, doc, func(n ast.Node) (ast.WalkStatus, error) {
		c := n.(*Citation)
		num, ok := citeNums[c.Key]
		if !ok {
			num = nextNum
			citeNums[c.Key] = num
			nextNum++
		}
		cnt := citeCounts[c.Key]
		citeCounts[c.Key] += 1

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
		c.Display = "[" + strconv.Itoa(num) + "]"
		c.ID = c.CiteID(cnt)
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
	w.WriteString(c.Display)
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
	w.WriteString(c.Display)
	w.WriteString(`</cite> `)

	renderCiteRefContent(w, c)
	w.WriteString(`</div>`)
}

func renderCiteRefContent(w util.BufWriter, c *Citation) {
	authors := c.Bibtex.Author
	for i, author := range authors {
		sp := strings.Split(author.First, " ")
		for _, s := range sp {
			if r, _ := utf8.DecodeRuneInString(s); r != utf8.RuneError {
				w.WriteRune(r)
				w.WriteString(". ")
			}
		}
		_, _ = w.WriteString(author.Last)
		if i < len(authors)-2 {
			w.WriteString(", ")
		} else if i == len(authors)-2 {
			if authors[len(authors)-1].IsOthers() {
				w.WriteString(" <em>et al</em>")
				break

			} else {
				w.WriteString(" and ")
			}
		}
	}

	title := c.Bibtex.Tags["title"]
	title = trimBraces(title)
	w.WriteString(`, "`)
	w.WriteString(title)
	w.WriteString(`,"`)

	hasInfoAfterTitle := false

	journal := c.Bibtex.Tags["journal"]
	journal = trimBraces(journal)
	if journal != "" {
		w.WriteString(" in <em class=cite-journal>")
		w.WriteString(journal)
		w.WriteString("</em>")
		hasInfoAfterTitle = true
	}

	vol := c.Bibtex.Tags["volume"]
	vol = trimBraces(vol)
	if vol != "" {
		if hasInfoAfterTitle {
			w.WriteRune(',')
		}
		w.WriteString(" Vol. ")
		w.WriteString(vol)
		hasInfoAfterTitle = true
	}

	year := c.Bibtex.Tags["year"]
	year = trimBraces(year)
	if year != "" {
		if hasInfoAfterTitle {
			w.WriteRune(',')
		}
		w.WriteRune(' ')
		w.WriteString(year)
	}

	w.WriteString(".")
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
