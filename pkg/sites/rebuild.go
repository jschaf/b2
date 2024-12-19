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

// Rebuild rebuilds everything on the site into dir.
func Rebuild(pubDir string) error {
	start := time.Now()

	if err := dirs.CleanDir(pubDir); err != nil {
		return fmt.Errorf("failed to clean public dir: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		slog.Debug("rebuild compile post details")
		c := compiler.NewDetailCompiler(pubDir)
		if err := c.Compile(""); err != nil {
			return fmt.Errorf("compile all detail posts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild compile root index")
		ic := compiler.NewIndexCompiler(pubDir)
		if err := ic.Compile(); err != nil {
			return fmt.Errorf("compile main index: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild copy all css")
		if _, err := css.CopyAllCSS(pubDir); err != nil {
			return fmt.Errorf("copy all css: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild copy all fonts")
		if err := css.CopyAllFonts(pubDir); err != nil {
			return fmt.Errorf("copy all fonts: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild copy static files")
		if err := static.CopyStaticFiles(pubDir); err != nil {
			return fmt.Errorf("copy static files: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild link papers")
		if err := static.LinkPapers(pubDir); err != nil {
			return fmt.Errorf("link papers: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Debug("rebuild typescript")
		if err := js.WriteTypeScriptMain(pubDir); err != nil {
			return fmt.Errorf("write typescript bundle: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("rebuild wait err group: %w", err)
	}

	slog.Info("rebuilt site", "duration", time.Since(start))
	return nil
}
