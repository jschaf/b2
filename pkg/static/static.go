package static

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jschaf/jsc/pkg/dirs"
	"github.com/jschaf/jsc/pkg/git"
	"github.com/jschaf/jsc/pkg/paths"
)

// CopyStaticFiles copies static files from the source static dir into
// distDir/static.
func CopyStaticFiles(distDir string) error {
	dir := git.RootDir()
	staticDir := filepath.Join(dir, dirs.Static)
	err := filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		// Skip .ts files because we bundle into JavaScript.
		if info.IsDir() || filepath.Ext(info.Name()) == ".ts" {
			return nil
		}
		rel, err := filepath.Rel(staticDir, path)
		if err != nil {
			return fmt.Errorf("failed to get rel path for static files: %w", err)
		}
		dest := filepath.Join(distDir, rel)
		return paths.Copy(dest, path)
	})
	if err != nil {
		return fmt.Errorf("failed to copy static files: %w", err)
	}
	return nil
}

// LinkPapers symlinks academic papers from the source papers dir into
// distDir/papers.
func LinkPapers(distDir string) error {
	dir := git.RootDir()
	papersDir := filepath.Join(dir, dirs.Papers)
	pubPapersDir := filepath.Join(distDir, "papers")
	err := os.Symlink(papersDir, pubPapersDir)
	if err != nil && !errors.Is(err.(*os.LinkError).Unwrap(), os.ErrExist) {
		return fmt.Errorf("link papers symlink: %w", err.(*os.LinkError).Unwrap())
	}
	return nil
}
