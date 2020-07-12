package sites

import (
	"fmt"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/static"
	"go.uber.org/zap"
	"time"
)

// Rebuild rebuilds everything on the site.
func Rebuild(l *zap.Logger) error {
	start := time.Now()

	c := compiler.NewForPostDetail(l)
	if err := c.CompileAllPosts(""); err != nil {
		return fmt.Errorf("compile all detail posts: %w", err)
	}

	ic := compiler.NewForIndex(l)
	if err := ic.Compile(); err != nil {
		return fmt.Errorf("compile main index: %w", err)
	}

	if err := static.CopyStaticFiles(); err != nil {
		return fmt.Errorf("copy static files: %w", err)
	}

	if err := static.LinkPapers(); err != nil {
		return fmt.Errorf("link papers: %w", err)
	}
	l.Sugar().Infof("rebuilt site in %.3f seconds", time.Since(start).Seconds())
	return nil
}
