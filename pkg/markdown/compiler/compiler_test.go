package compiler

import (
	"go.uber.org/zap"
	"testing"
)

func BenchmarkCompiler_CompileAllPosts(b *testing.B) {
	b.StopTimer()
	c := NewForPostDetail(zap.NewNop())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err := c.CompileAllPosts("procella"); err != nil {
			b.Fatal(err)
		}
	}
}
