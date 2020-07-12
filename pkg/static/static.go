package static

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/paths"
)

func CopyStaticFiles() error {
	dir, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("failed to find root dir for static files: %w", err)
	}
	staticDir := filepath.Join(dir, dirs.Static)
	pubDir := filepath.Join(dir, dirs.Public)
	err = filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(staticDir, path)
		if err != nil {
			return fmt.Errorf("failed to get rel path for static files: %w", err)
		}
		dest := filepath.Join(pubDir, rel)
		return paths.Copy(path, dest)
	})

	if err != nil {
		return fmt.Errorf("failed to copy static files: %w", err)
	}
	return nil
}

func LinkPapers() error {
	dir, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("LinkPapers find root dir: %w", err)
	}
	papersDir := filepath.Join(dir, dirs.Papers)
	pubPapersDir := filepath.Join(dir, dirs.Public, "papers")
	if err := os.Symlink(papersDir, pubPapersDir); err != nil {
		return fmt.Errorf("LinkPapers symlink: %w", err)
	}
	return nil
}
