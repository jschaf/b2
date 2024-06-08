package mdext

import (
	"testing"

	"github.com/jschaf/b2/pkg/markdown/mdtest"

	"github.com/jschaf/b2/pkg/htmls/tags"
)

func TestSmallCapsExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"foo", "foo"},
		{"FO", "FO"},
		{"FOO", tags.SC("FOO")},
		{"SS0", tags.SC("SS0")},
		{"(FOO)", tags.SC("(FOO)")},
		{"(FOO.", "(" + tags.SC("FOO") + "."},
		{"FOO)", tags.SC("FOO") + ")"},
		{"FOO,", tags.SC("FOO") + ","},
		{"FOOs", tags.SC("FOO") + "s"},
		{"FOOss", `FOOss`},
		{"F_BAR", `F_BAR`},
		{"FOO_BAR", `FOO_BAR`},
		{"MOTD\n", tags.SC("MOTD")},
		{"alpha MOTD\nfoo", "alpha " + tags.SC("MOTD") + " foo"},
		{"*FOO*", tags.Em(tags.SC("FOO"))},
		{"**FOO**", tags.Strong(tags.SC("FOO"))},
		{"FOO", tags.SC("FOO")},
		{"STUBBLE", tags.SC("STUBBLE")},
		{"FOO BAR", tags.SC("FOO") + " " + tags.SC("BAR")},
		{"The (MOTD)", "The " + tags.SC("(MOTD)")},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewSmallCapsExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tags.P(tt.want))
		})
	}
}
