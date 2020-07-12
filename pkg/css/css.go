package css

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/paths"
)

// WriteMainCSS writes the main CSS stylesheet.
func WriteMainCSS(root string) (string, error) {
	dest := filepath.Join(root, dirs.Public, "style", "main.css")
	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create dir public/style: %w", err)
	}
	err = paths.Copy(filepath.Join(root, dirs.Style, "main.css"), dest)
	if err != nil {
		return "", fmt.Errorf("copy main.css to public: %w", err)
	}
	return dest, nil
}
