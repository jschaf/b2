package compiler

import (
	"github.com/jschaf/b2/pkg/dirs"
	"go.uber.org/zap"
	"testing"
)

func BenchmarkCompiler_CompileAllPosts(b *testing.B) {
	b.StopTimer()
	c := NewForPostDetail(dirs.PublicMemfs, zap.NewNop())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err := c.CompileAllPosts("procella"); err != nil {
			b.Fatal(err)
		}
	}
}
