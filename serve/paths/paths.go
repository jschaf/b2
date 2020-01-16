package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	rootOnce sync.Once
	rootDir  string
	rootErr  error
)

// findRootDir finds the root directory which is the nearest directory
// containing a .git folder.
func FindRootDir() (string, error) {
	rootOnce.Do(func() {
		rootDir, rootErr = walkUp(".git")
	})
	return rootDir, rootErr
}

func walkUp(dirToFind string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working dir: %s", err)
	}

	for dir != string(os.PathSeparator) {
		p := filepath.Join(dir, dirToFind)

		if stat, err := os.Stat(p); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("failed to stat %s: %w", p, err)
			}
		} else if stat.IsDir() {
			return dir, nil
		}

		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("git dir not found starting from %s", dir)
}
