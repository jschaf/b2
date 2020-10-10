package css

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/karrick/godirwalk"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jschaf/b2/pkg/paths"
)

// CopyAllCSS copies all CSS files into pubDir/style.
func CopyAllCSS(pubDir string) ([]string, error) {
	styleDir := filepath.Join(git.MustFindRootDir(), dirs.Style)
	destDir := filepath.Join(pubDir, dirs.Style)
	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return nil, fmt.Errorf("create public style dir: %w", err)
	}

	cb := func(path string, dirent *godirwalk.Dirent) ([]string, error) {
		if !dirent.IsRegular() || filepath.Ext(path) != ".css" {
			return nil, nil
		}
		rel, err := filepath.Rel(styleDir, path)
		if err != nil {
			return nil, fmt.Errorf("rel path for css: %w", err)
		}
		dest := filepath.Join(destDir, rel)
		if isSame, err := paths.CopyLazy(dest, path); err != nil {
			return nil, fmt.Errorf("copy lazy css file: %w", err)
		} else if isSame {
			return []string{dest}, nil
		}
		return nil, nil
	}
	cssPaths, err := paths.WalkCollectStrings(styleDir, runtime.NumCPU(), cb)
	if err != nil {
		return nil, fmt.Errorf("copy css to public dir: %w", err)
	}
	return cssPaths, nil
}

// CopyAllFonts copies all font files into pubDir/fonts.
func CopyAllFonts(pubDir string) error {
	fontDir := filepath.Join(git.MustFindRootDir(), dirs.Style, dirs.Fonts)
	destDir := filepath.Join(pubDir, dirs.Style, dirs.Fonts)
	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return fmt.Errorf("create public font dir: %w", err)
	}

	cb := func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".woff2" {
			return nil
		}
		rel, err := filepath.Rel(fontDir, path)
		if err != nil {
			return fmt.Errorf("rel path for font: %w", err)
		}
		dest := filepath.Join(destDir, rel)
		if _, err := paths.CopyLazy(dest, path); err != nil {
			return fmt.Errorf("copy lazy font file: %w", err)
		}
		return nil
	}
	err := paths.WalkConcurrent(fontDir, runtime.NumCPU(), cb)
	if err != nil {
		return fmt.Errorf("copy css to public dir: %w", err)
	}
	return nil
}
