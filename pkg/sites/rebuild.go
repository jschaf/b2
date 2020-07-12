package sites

import (
	"context"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/static"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"time"
)

// Rebuild rebuilds everything on the site.
func Rebuild(l *zap.Logger) error {
	start := time.Now()

	if err := dirs.CleanPubDir(); err != nil {
		return fmt.Errorf("failed to clean public dir: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		c := compiler.NewForPostDetail(l)
		if err := c.CompileAllPosts(""); err != nil {
			return fmt.Errorf("compile all detail posts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		ic := compiler.NewForIndex(l)
		if err := ic.Compile(); err != nil {
			return fmt.Errorf("compile main index: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := static.CopyStaticFiles(); err != nil {
			return fmt.Errorf("copy static files: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := static.LinkPapers(); err != nil {
			return fmt.Errorf("link papers: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("rebuild wait err group: %w", err)
	}

	l.Sugar().Infof("rebuilt site in %.3f seconds", time.Since(start).Seconds())
	return nil
}
