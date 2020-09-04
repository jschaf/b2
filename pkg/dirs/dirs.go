package dirs

import (
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"os"
	"path/filepath"
)

const (
	// Top level directories.
	Cmd         = "cmd"
	Papers      = "papers"
	Pkg         = "pkg"
	Posts       = "posts"
	Public      = "public"
	Scripts     = "scripts"
	Static      = "static"
	Style       = "style"
	TIL       = "til"
	PublicMemfs = "/m/b2" // alternate public dir for development
)

// RemoveAllChildren removes all children in the directory.
func RemoveAllChildren(dir string) (mErr error) {
	f, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open dir: %w", err)
	}
	defer errs.CapturingClose(&mErr, f, "close readdir")

	files, err := f.Readdir(-1)
	if err != nil {
		return fmt.Errorf("readdir: %w", err)
	}

	multiErr := errs.NewMultiError()
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		err := os.RemoveAll(path)
		multiErr.Append(err)
	}
	return multiErr.ErrorOrNil()
}

// CleanDir creates dir if it doesn't exist and then deletes all children of the
// dir.
func CleanDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("make dir: %w", err)
	}
	if err := RemoveAllChildren(dir); err != nil {
		return fmt.Errorf("remove all dir children: %w", err)
	}
	return nil
}
