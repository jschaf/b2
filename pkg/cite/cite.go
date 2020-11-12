package cite

import "github.com/jschaf/bibtex"

type Style string

const (
	IEEE Style = "IEEE"
)

// Biber controls parsing, resolving, and rendering of bibtex used in side notes
// and end notes.
var Biber = bibtex.New(
	bibtex.WithResolvers(
		bibtex.NewAuthorResolver("author"),
		bibtex.ResolverFunc(bibtex.SimplifyEscapedTextResolver),
		bibtex.NewRenderParsedTextResolver(),
	))
