package mdext

import (
	"bytes"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"regexp"
	"strings"
)

var superscriptRegexp = regexp.MustCompile(`\^[a-z0-9]`)

func renderTextTitle(reader text.Reader, heading *ast.Heading) string {
	// Collect raw source since the Text() method is deprecated. The Katex
	// extension doesn't provide it as well.
	buf := &bytes.Buffer{}
	lines := heading.Lines()
	for i := range lines.Len() {
		seg := heading.Lines().At(i)
		buf.Write(seg.Value(reader.Source()))
	}
	s := buf.String()

	// Assume there's only a single math element in the title.
	lo := strings.IndexByte(s, '$')
	if lo == -1 {
		return string(heading.Text(reader.Source()))
	}
	hi := strings.IndexByte(s[lo+1:], '$')
	if hi == -1 {
		return string(heading.Text(reader.Source()))
	}
	hi += lo + 1

	before := s[:lo]
	math := s[lo+1 : hi]
	after := s[hi+1:]

	// Replace superscript numbers with actual superscript.
	math = superscriptRegexp.ReplaceAllStringFunc(math, func(s string) string {
		return toSuperscript(strings.TrimPrefix(s, "^"))
	})

	return before + math + after
}

func toSuperscript(s string) string {
	sb := strings.Builder{}
	for _, r := range s {
		switch r {
		case '0':
			sb.WriteRune('⁰')
		case '1':
			sb.WriteRune('¹')
		case '2':
			sb.WriteRune('²')
		case '3':
			sb.WriteRune('³')
		case '4':
			sb.WriteRune('⁴')
		case '5':
			sb.WriteRune('⁵')
		case '6':
			sb.WriteRune('⁶')
		case '7':
			sb.WriteRune('⁷')
		case '8':
			sb.WriteRune('⁸')
		case '9':
			sb.WriteRune('⁹')
		case 'a':
			sb.WriteRune('ᵃ')
		case 'b':
			sb.WriteRune('ᵇ')
		case 'c':
			sb.WriteRune('ᶜ')
		case 'd':
			sb.WriteRune('ᵈ')
		case 'e':
			sb.WriteRune('ᵉ')
		case 'f':
			sb.WriteRune('ᶠ')
		case 'g':
			sb.WriteRune('ᵍ')
		case 'h':
			sb.WriteRune('ʰ')
		case 'i':
			sb.WriteRune('ⁱ')
		case 'j':
			sb.WriteRune('ʲ')
		case 'k':
			sb.WriteRune('ᵏ')
		case 'l':
			sb.WriteRune('ˡ')
		case 'm':
			sb.WriteRune('ᵐ')
		case 'n':
			sb.WriteRune('ⁿ')
		case 'o':
			sb.WriteRune('ᵒ')
		case 'p':
			sb.WriteRune('ᵖ')
		case 'r':
			sb.WriteRune('ʳ')
		case 's':
			sb.WriteRune('ˢ')
		case 't':
			sb.WriteRune('ᵗ')
		case 'u':
			sb.WriteRune('ᵘ')
		case 'v':
			sb.WriteRune('ᵛ')
		case 'w':
			sb.WriteRune('ʷ')
		case 'x':
			sb.WriteRune('ˣ')
		case 'y':
			sb.WriteRune('ʸ')
		case 'z':
			sb.WriteRune('ᶻ')
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
