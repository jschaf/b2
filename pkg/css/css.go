package css

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/paths"
)

// WriteMainCSS writes the main CSS stylesheet into pubDir.
func WriteMainCSS(pubDir string) (string, error) {
	src := filepath.Join(git.MustFindRootDir(), dirs.Style, "main.css")
	dest := filepath.Join(pubDir, "style", "main.css")
	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create dir public/style: %w", err)
	}
	err = paths.Copy(dest, src)
	if err != nil {
		return "", fmt.Errorf("copy main.css to public: %w", err)
	}
	return dest, nil
}
