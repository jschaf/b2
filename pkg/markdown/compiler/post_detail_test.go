package compiler

import (
	"testing"

	"github.com/jschaf/b2/pkg/dirs"
)

func BenchmarkCompiler_CompileAllPosts(b *testing.B) {
	b.StopTimer()
	c := NewPostDetail(dirs.PublicMemfs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err := c.CompileAll("procella"); err != nil {
			b.Fatal(err)
		}
	}
}
