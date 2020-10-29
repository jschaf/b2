package mdext

import "C"
import (
	"fmt"
	"github.com/jschaf/b2/pkg/bibtex"
	"github.com/jschaf/b2/pkg/texts"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

type absPath = string

// citationRendererIEEE renders an IEEE citation.
//
// All data in the renderer uses map with a key of the AbsPath. We need this
// because we retain state in this struct. That state gets reused for other
// posts. We don't have parser.Context which was scoped to a single document.
type citationRendererIEEE struct {
	postStates map[absPath]*ieeeState
}

func (cr *citationRendererIEEE) getPostState(path absPath) *ieeeState {
	if p, ok := cr.postStates[path]; ok {
		return p
	}
	st := &ieeeState{
		nextNum:    1,
		citeNums:   map[bibtex.CiteKey]int{},
		citeCounts: make(map[bibtex.CiteKey]int),
	}
	cr.postStates[path] = st
	return st
}

// ieeeState is the citation state for a single post.
type ieeeState struct {
	// Next number to use as a citation reference, like [1]. Starts at 1.
	nextNum int
	// Mapping from the abs path to the bibtex cite key to the order the first
	// instance of a cite key appeared in the markdown document.
	citeNums map[bibtex.CiteKey]int
	// The number of times a bibtex cite key has been used thus far. Useful for
	// generating unique IDs for the citation.
	citeCounts map[bibtex.CiteKey]int
}

func (cr *citationRendererIEEE) renderCitation(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	c := n.(*Citation)
	// For IEEE style we need to dedupe the citation order. The raw order
	// assigns multiple numbers for the same cite key.
	st := cr.getPostState(c.AbsPath)
	num, ok := st.citeNums[c.Key]
	if !ok {
		num = st.nextNum
		st.citeNums[c.Key] = num
		st.nextNum += 1
	}

	cnt := st.citeCounts[c.Key]
	st.citeCounts[c.Key] = cnt + 1

	w.WriteString(
		fmt.Sprintf(`<a href="%s/#%s" class=preview-target data-link-type=%s>`,
			c.AbsPath, c.ReferenceID(), LinkCitation))

	id := c.CiteID(cnt)
	_, _ = w.WriteString(fmt.Sprintf(`<cite id=%s>[%d]</cite>`, id, num))
	_, _ = w.WriteString("</a>")
	// Citations should generate content solely from the citation, not children.
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

	_, _ = w.WriteString(`<div class=cite-references>`)
	_, _ = w.WriteString(`<h2>References</h2>`)
	for _, c := range refs.Citations {
		if _, ok := hasRef[c.Key]; ok {
			continue
		}
		hasRef[c.Key] = struct{}{}
		num, ok := cr.getPostState(c.AbsPath).citeNums[c.Key]
		if !ok {
			return ast.WalkStop, fmt.Errorf("citation: no number found for reference: %s", c.Key)
		}
		cr.renderCiteRef(w, c, num)
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

func (cr *citationRendererIEEE) renderCiteRef(w util.BufWriter, c *Citation, num int) {
	cnt := cr.getPostState(c.AbsPath).citeCounts[c.Key]
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
	w.WriteString(`">[`)
	w.WriteString(strconv.Itoa(num))
	w.WriteString(`]</cite> `)

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
	w.WriteString(`</div>`)
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
