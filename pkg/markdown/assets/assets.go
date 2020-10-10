package assets

import (
	"fmt"
	"github.com/jschaf/b2/pkg/paths"
	"path/filepath"
)

// Map maps from the relative URL to the full file path of an asset like an
// image. For example, 1 entry might be ./img.png -> /home/joe/blog/img.png.
type Map = map[string]string

// CopyAll copies all assets into the pubDir, overwriting existing files.
func CopyAll(pubDir string, assets Map) error {
	for destPath, srcPath := range assets {
		dest := filepath.Join(pubDir, destPath)
		if _, err := paths.CopyLazy(dest, srcPath); err != nil {
			return fmt.Errorf("failed to copy asset to dest: %w", err)
		}
	}
	return nil
}
