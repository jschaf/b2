package sites

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jschaf/b2/pkg/css"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/js"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/static"
	"golang.org/x/sync/errgroup"
)

// Rebuild rebuilds everything on the site into distDir.
func Rebuild(distDir string) error {
	slog.Info("start rebuild site")
	start := time.Now()

	if err := dirs.CleanDir(distDir); err != nil {
		return fmt.Errorf("failed to clean public dir: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		slog.Debug("rebuild compile details")
		c := compiler.NewDetailCompiler(distDir)
		if err := c.Compile(""); err != nil {
			return fmt.Errorf("compile all detail posts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild compile index")
		ic := compiler.NewIndexCompiler(distDir)
		if err := ic.Compile(); err != nil {
			return fmt.Errorf("compile main index: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild copy all css")
		if _, err := css.CopyAllCSS(distDir); err != nil {
			return fmt.Errorf("copy all css: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild copy all fonts")
		if err := css.CopyAllFonts(distDir); err != nil {
			return fmt.Errorf("copy all fonts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild copy static files")
		if err := static.CopyStaticFiles(distDir); err != nil {
			return fmt.Errorf("copy static files: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild link papers")
		if err := static.LinkPapers(distDir); err != nil {
			return fmt.Errorf("link papers: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild typescript")
		if err := js.WriteTypeScriptMain(distDir); err != nil {
			return fmt.Errorf("write typescript bundle: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("rebuild wait err group: %w", err)
	}

	slog.Info("finish rebuild site", "duration", time.Since(start))
	return nil
}
