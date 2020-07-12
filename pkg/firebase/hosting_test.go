package firebase

import (
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestSiteHashes_PopulateFromDir(t *testing.T) {
	tests := []struct {
		dir             string
		wantHashesByUrl map[string]string
	}{
		{"testdata/site_hashes", map[string]string{
			"/bar.html":               "0d9630c0f315aad6134273e7c52bb660d897275325ef1ebab7c3fa88831d2a34",
			"/foo.html":               "9a1b21db16d717ebebcc60ab1bf892be8870f777a65dd50394b4194d9bc13124",
			"/linked_site/circle.png": "d4b58b6388b662ef70ff13152acbf3e3e5a539ffe5855027fc1aa117738d9d90",
			"/qux.html":               "f3b8722118a8ac3a4e226d2b6e220e5bd7425c162b8c207c0fa529e1aca5ac10",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			sh := NewSiteHashes(zaptest.NewLogger(t).Sugar())
			if err := sh.PopulateFromDir(tt.dir); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wantHashesByUrl, sh.HashesByURL()); diff != "" {
				t.Errorf("HashesByURL() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkSiteHashes_PopulateFromDir(b *testing.B) {
	b.StopTimer()
	l, _ := zap.NewDevelopment()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		sh := NewSiteHashes(l.Sugar())
		if err := sh.PopulateFromDir("../../public"); err != nil {
			b.Fatal(err)
		}
	}
}
