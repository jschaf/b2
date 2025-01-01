package assets

import (
	"fmt"
	"path/filepath"

	"github.com/jschaf/jsc/pkg/paths"
)

type Blob struct {
	// Absolute path of the source file. If nil, GenPath must be non-nil.
	Src string
	// Path relative to the pub dir of the destination file path.
	Dest string
	// If non-nil, how to generate the output for Dest.
	GenFunc func() error
}

// CopyAll copies all assets into the distDir, overwriting existing files.
func CopyAll(distDir string, assets []Blob) error {
	for _, blob := range assets {
		dest := filepath.Join(distDir, blob.Dest)
		if _, err := paths.CopyLazy(dest, blob.Src); err != nil {
			return fmt.Errorf("failed to copy asset to dest: %w", err)
		}
	}
	return nil
}
