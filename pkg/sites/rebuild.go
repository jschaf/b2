package sites

import (
	"context"
	"fmt"
	"github.com/jschaf/b2/pkg/css"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/static"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"time"
)

// Rebuild rebuilds everything on the site into dir.
func Rebuild(pubDir string, l *zap.Logger) error {
	start := time.Now()

	if err := dirs.CleanDir(pubDir); err != nil {
		return fmt.Errorf("failed to clean public dir: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		c := compiler.NewForPostDetail(pubDir, l)
		if err := c.CompileAllPosts(""); err != nil {
			return fmt.Errorf("compile all detail posts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		ic := compiler.NewForIndex(pubDir, l)
		if err := ic.Compile(); err != nil {
			return fmt.Errorf("compile main index: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		tc := compiler.NewForTIL(pubDir, l)
		if err := tc.CompileAllTILs(); err != nil {
			return fmt.Errorf("compile all TILs: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if _, err := css.WriteMainCSS(pubDir); err != nil {
			return fmt.Errorf("write main.css: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := static.CopyStaticFiles(pubDir); err != nil {
			return fmt.Errorf("copy static files: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := static.LinkPapers(pubDir); err != nil {
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
