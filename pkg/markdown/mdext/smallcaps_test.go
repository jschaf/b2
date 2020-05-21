package mdext

import (
	"testing"
)

func TestSmallCapsExt(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"foo", "<p>foo</p>"},
		{"FO", "<p>FO</p>"},
		{"FOO", `<p><span class="small-caps">FOO</span></p>`},
		{"(FOO)", `<p><span class="small-caps">(FOO)</span></p>`},
		{"(FOO.", `<p>(<span class="small-caps">FOO</span>.</p>`},
		{"FOO)", `<p><span class="small-caps">FOO</span>)</p>`},
		{"F_BAR", `<p>F_BAR</p>`},
		{"FOO_BAR", `<p>FOO_BAR</p>`},
		{"MOTD\n", `<p><span class="small-caps">MOTD</span></p>`},
		{"alpha MOTD\nfoo", "<p>alpha <span class=\"small-caps\">MOTD</span> foo</p>"},
		{"*FOO*", `<p><em><span class="small-caps">FOO</span></em></p>`},
		{"**FOO**", `<p><strong><span class="small-caps">FOO</span></strong></p>`},
		{"FOO", `<p><span class="small-caps">FOO</span></p>`},
		{"STUBBLE", `<p><span class="small-caps">STUBBLE</span></p>`},
		{"FOO BAR", `<p><span class="small-caps">FOO</span> <span class="small-caps">BAR</span></p>`},
		{"The (MOTD)", `<p>The <span class="small-caps">(MOTD)</span>`},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			md, ctx := newMdTester(t, NewSmallCapsExt())
			assertNoRenderDiff(t, md, ctx, tt.src, tt.want)
		})
	}
}
