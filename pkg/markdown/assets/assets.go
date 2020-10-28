package assets

import (
	"fmt"
	"github.com/jschaf/b2/pkg/paths"
	"path/filepath"
)

type Blob struct {
	// Absolute path of the source file. If nil, GenPath must be non-nil.
	Src string
	// Path relative to the pub dir of the destination file path.
	Dest string
	// If non-nil, how to generate the output for Dest.
	GenFunc func() error
}

// CopyAll copies all assets into the pubDir, overwriting existing files.
func CopyAll(pubDir string, assets []Blob) error {
	for _, blob := range assets {
		dest := filepath.Join(pubDir, blob.Dest)
		if _, err := paths.CopyLazy(dest, blob.Src); err != nil {
			return fmt.Errorf("failed to copy asset to dest: %w", err)
		}
	}
	return nil
}
