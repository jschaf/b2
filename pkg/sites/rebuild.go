package sites

import (
	"context"
	"fmt"
	"github.com/jschaf/b2/pkg/css"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/js"
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
		l.Debug("Rebuild - compile post details")
		c := compiler.NewPostDetail(pubDir, l)
		if err := c.CompileAll(""); err != nil {
			return fmt.Errorf("compile all detail posts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - compile root index")
		ic := compiler.NewRootIndex(pubDir, l)
		if err := ic.CompileIndex(); err != nil {
			return fmt.Errorf("compile main index: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - compile TIL details")
		tc := compiler.NewTILDetail(pubDir, l)
		if err := tc.CompileAll(); err != nil {
			return fmt.Errorf("compile all TIL posts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - compile book details")
		c := compiler.NewBookDetail(pubDir, l)
		if err := c.CompileAll(); err != nil {
			return fmt.Errorf("compile all book details: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - compile TIL index")
		tc := compiler.NewTILIndex(pubDir, l)
		if err := tc.CompileIndex(); err != nil {
			return fmt.Errorf("compile TIL index: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - copy all CSS")
		if _, err := css.CopyAllCSS(pubDir); err != nil {
			return fmt.Errorf("copy all css: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - Copy all fonts")
		if err := css.CopyAllFonts(pubDir); err != nil {
			return fmt.Errorf("copy all fonts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - copy static files")
		if err := static.CopyStaticFiles(pubDir); err != nil {
			return fmt.Errorf("copy static files: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - link papers")
		if err := static.LinkPapers(pubDir); err != nil {
			return fmt.Errorf("link papers: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		l.Debug("Rebuild - typescript")
		if err := js.WriteTypeScriptMain(pubDir); err != nil {
			return fmt.Errorf("write typescript bundle: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("rebuild wait err group: %w", err)
	}

	l.Sugar().Infof("rebuilt site in %.3f seconds", time.Since(start).Seconds())
	return nil
}
