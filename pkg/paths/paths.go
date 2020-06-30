package paths

import (
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"io"
	"os"
	"path/filepath"
)

// WalkUp traverses up directory tree until it finds an ancestor directory that
// contains dirToFind. WalkUp checks the current directory and then
func WalkUp(dirToFind string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working dir: %s", err)
	}

	for dir != string(os.PathSeparator) {
		p := filepath.Join(dir, dirToFind)

		if stat, err := os.Stat(p); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("stat dirToFind %s: %w", p, err)
			}
		} else if stat.IsDir() {
			return dir, nil
		}

		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("dir not found in directory tree starting from %s", dir)
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Copy(src, dst string) (mErr error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer errs.CloseWithErrCapture(&mErr, in, "")

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer errs.CloseWithErrCapture(&mErr, out, "")

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}
