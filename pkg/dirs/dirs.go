package dirs

import (
	"errors"
	"fmt"
	"github.com/jschaf/b2/pkg/git"
	"os"
	"path/filepath"
)

// Top level directories.
const (
	Cmd     = "cmd"
	Papers  = "papers"
	Pkg     = "pkg"
	Posts   = "posts"
	Public  = "public"
	Scripts = "scripts"
	Static  = "static"
	Style   = "style"
)

func CleanPubDir() error {
	publicDir := filepath.Join(git.MustFindRootDir(), Public)

	if stat, err := os.Stat(publicDir); err == nil {
		if !stat.IsDir() {
			return errors.New("public dir is not a directory")
		}
		if err := os.RemoveAll(publicDir); err != nil {
			return fmt.Errorf("failed to delete public dir: %w", err)
		}
	} else if os.IsNotExist(err) {
		// Do nothing.
	} else {
		return fmt.Errorf("failed to stat pub dir: %w", err)
	}

	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return fmt.Errorf("failed to make public dir: %w", err)
	}
	return nil
}
