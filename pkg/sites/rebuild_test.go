package sites

import (
	"testing"

	"github.com/jschaf/b2/pkg/dirs"
)

func BenchmarkRebuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Rebuild(dirs.Public); err != nil {
			b.Fatal(err)
		}
	}
}
