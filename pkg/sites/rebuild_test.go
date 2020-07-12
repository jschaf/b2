package sites

import (
	"go.uber.org/zap"
	"testing"
)

func BenchmarkRebuild(b *testing.B) {
	l := zap.NewNop()
	for i := 0; i < b.N; i++ {
		if err := Rebuild(l); err != nil {
			b.Fatal(err)
		}
	}
}
