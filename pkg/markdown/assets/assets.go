package assets

import (
	"errors"
	"fmt"
	"github.com/jschaf/b2/pkg/files"
	"github.com/jschaf/b2/pkg/paths"
	"os"
	"path/filepath"
)

// Map maps from the relative URL to the full file path of an asset like an
// image. For example, 1 entry might be ./img.png -> /home/joe/blog/img.png.
type Map = map[string]string

// CopyAll copies all assets into the pubDir, overwriting existing files.
func CopyAll(pubDir string, assets Map) error {
	for destPath, srcPath := range assets {
		dest := filepath.Join(pubDir, destPath)
		if isSame, err := files.SameBytes(srcPath, dest); errors.Is(err, os.ErrNotExist) {
			// Ignore
		} else if err != nil {
			return fmt.Errorf("check if file contents are same: %w", err)
		} else if isSame {
			continue
		}

		if err := paths.Copy(dest, srcPath); err != nil {
			return fmt.Errorf("failed to copy asset to dest: %w", err)
		}
	}
	return nil
}
