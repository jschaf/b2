package sites

import (
	"github.com/jschaf/b2/pkg/dirs"
	"go.uber.org/zap"
	"testing"
)

func BenchmarkRebuild(b *testing.B) {
	l := zap.NewNop()
	for i := 0; i < b.N; i++ {
		if err := Rebuild(dirs.PublicMemfs, l); err != nil {
			b.Fatal(err)
		}
	}
}
