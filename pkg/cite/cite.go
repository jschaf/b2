package cite

import (
	"strings"

	"github.com/jschaf/b2/pkg/cite/bibtex"
)

type Style string

const (
	IEEE Style = "IEEE"
)

// Names come mostly in 3 forms:
// - First von Last
// - von Last, First
// - von Last, Jr, First
// See https://nwalsh.com/tex/texhelp/bibtx-23.html
type Author struct {
	First string
	Von   string
	Last  string
	Jr    string
}

func ParseAuthors(b *bibtex.Element) []Author {
	a, ok := b.Tags["author"]
	if !ok {
		return nil
	}
	a = strings.Trim(a, `"{}`)
	as := strings.Split(a, " and ")
	ps := make([]Author, len(as))
	for i, author := range as {
		ps[i] = ParseAuthor(author)
	}
	return ps
}

func ParseAuthor(author string) Author {
	a := Author{}
	// TODO: Proper parsing
	a.Last = author
	return a
}
