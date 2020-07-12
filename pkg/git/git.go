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

// FindRootDir finds the nearest directory containing a .git folder. Checks
// the current dir and then walks up parent directories.
func FindRootDir() (string, error) {
	rootOnce.Do(func() {
		rootDir, rootErr = paths.WalkUp(".git")
	})
	return rootDir, rootErr
}

// FindRootDir finds the nearest directory containing a .git folder. Checks
// the current dir and then walks up parent directories. Panics if no parent
// directory contains a .git folder.
func MustFindRootDir() string {
	dir, err := FindRootDir()
	if err != nil {
		panic(err)
	}
	return dir
}
