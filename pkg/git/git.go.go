package git

import (
	"github.com/jschaf/b2/pkg/paths"
	"sync"
)

var (
	rootOnce sync.Once
	rootDir  string
	rootErr  error
)

// findRootDir finds the root directory which is the nearest parent directory
// containing a .git folder.
func FindRootDir() (string, error) {
	rootOnce.Do(func() {
		rootDir, rootErr = paths.WalkUp(".git")
	})
	return rootDir, rootErr
}
