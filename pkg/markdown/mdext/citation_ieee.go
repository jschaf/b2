package mdext

import "C"
import (
	"fmt"
	"github.com/jschaf/b2/pkg/bibtex"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// citationRendererIEEE renders an IEEE citation.
type citationRendererIEEE struct {
	nextNum  int
	citeNums map[bibtex.CiteKey]int
}

func (cr *citationRendererIEEE) renderCitation(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	c := n.(*Citation)
	// For IEEE style we need to dedupe the citation order. The raw order
	// assigns multiple numbers for the same cite key.
	num, ok := cr.citeNums[c.Key]
	if !ok {
		num = cr.nextNum
		cr.citeNums[c.Key] = num
		cr.nextNum += 1
	}

	_, _ = w.WriteString(
		fmt.Sprintf(`<a href="#%s" class=preview-target data-link-type=%s>`,
			c.ReferenceID(), LinkCitation))
	_, _ = w.WriteString(
		fmt.Sprintf(`<cite id=%s>[%d]</cite>`, c.ID(), num))
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
		num, ok := cr.citeNums[c.Key]
		if !ok {
			return ast.WalkStop, fmt.Errorf("citation: no number found for reference: %s", c.Key)
		}
		cr.renderCiteRef(w, c, num)
	}
	_, _ = w.WriteString(`</div>`)

	return ast.WalkContinue, nil
}

func (cr *citationRendererIEEE) renderCiteRef(w util.BufWriter, c *Citation, num int) {
	_, _ = w.WriteString(fmt.Sprintf(
		`<div id="%s" class=cite-reference>`, c.ReferenceID()))
	_, _ = w.WriteString(fmt.Sprintf(`<cite>[%d]</cite> `, num))

	authors := c.Bibtex.Author
	for i, author := range authors {
		sp := strings.Split(author.First, " ")
		for _, s := range sp {
			if r, _ := utf8.DecodeRuneInString(s); r != utf8.RuneError {
				_, _ = w.WriteRune(r)
				_, _ = w.WriteString(". ")
			}
		}
		_, _ = w.WriteString(author.Last)
		if i < len(authors)-2 {
			_, _ = w.WriteString(", ")
		} else if i == len(authors)-2 {
			if authors[len(authors)-1].IsOthers() {
				_, _ = w.WriteString(" <em>et al</em>")
				break

			} else {
				_, _ = w.WriteString(" and ")
			}

		}
	}

	title := c.Bibtex.Tags["title"]
	title = strings.Trim(title, `"{}`)
	_, _ = w.WriteString(fmt.Sprintf(`, "%s,"`, title))

	journal := c.Bibtex.Tags["journal"]
	journal = strings.Trim(journal, `"{}`)
	if journal != "" {
		_, _ = w.WriteString(fmt.Sprintf(" in <em class=cite-journal>%s</em>", journal))
	}

	vol := c.Bibtex.Tags["volume"]
	vol = strings.Trim(vol, `"{}`)
	if vol != "" {
		_, _ = w.WriteString(fmt.Sprintf(", Vol. %s", vol))
	}

	year := c.Bibtex.Tags["year"]
	year = strings.Trim(year, `"{}`)
	if year != "" {
		_, _ = w.WriteString(fmt.Sprintf(", %s", year))
	}

	_, _ = w.WriteString(".")
	_, _ = w.WriteString(`</div>`)
}
