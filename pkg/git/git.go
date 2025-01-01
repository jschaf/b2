package git

import (
	"sync"

	"github.com/jschaf/jsc/pkg/paths"
)

var rootFunc = sync.OnceValues[string, error](func() (string, error) {
	return paths.WalkUp(".git")
})

// RootDir finds the nearest directory containing a .git folder or panics.
// Checks the current dir and then walks up parent directories.
// Panics if no parent directory contains a .git folder.
func RootDir() string {
	dir, err := rootFunc()
	if err != nil {
		panic(err)
	}
	return dir
}
