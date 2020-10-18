package js

import (
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/files"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/jschaf/b2/pkg/git"
)

type jsCache struct {
	mu           sync.Mutex
	key          uint64
	bundleResult api.BuildResult
}

// Single entry cache for main.js.
var mainJSCache = &jsCache{}

func (jsCache *jsCache) isUnchanged(mainJSPath string) (bool, uint64, error) {
	newKey, err := files.HashFnv64(mainJSPath)
	if err != nil {
		return false, 0, fmt.Errorf("hash main.js for JS build cache: %w", err)
	}
	curKey := mainJSCache.key
	if newKey == curKey {
		return true, newKey, nil
	}
	return false, newKey, nil
}

// bundleMain minifies the main.js file and returns the bytes of the written file.
func bundleMain(pubDir string) (api.BuildResult, error) {
	mainJS := filepath.Join(git.MustFindRootDir(), dirs.Scripts, "main.js")
	mainJSOut := filepath.Join(pubDir, "main.js")

	// Check if file is same and skip JS bundle.
	mainJSCache.mu.Lock()
	defer mainJSCache.mu.Unlock()
	ok, newKey, err := mainJSCache.isUnchanged(mainJS)
	if err != nil {
		return api.BuildResult{}, err
	} else if ok {
		return mainJSCache.bundleResult, nil
	}

	// Continue with a full build.
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{mainJS},
		Outfile:     mainJSOut,
		Bundle:      false,
		Write:       true,
		ErrorLimit:  3,
		LogLevel:    api.LogLevelInfo,
	})

	mainJSCache.key = newKey
	mainJSCache.bundleResult = result
	return result, nil
}

func WriteMainBundle(pubDir string) error {
	result, err := bundleMain(pubDir)
	if err != nil {
		return fmt.Errorf("write main.js bundle: %w", err)
	}
	for _, f := range result.OutputFiles {
		if err := ioutil.WriteFile(f.Path, f.Contents, 0644); err != nil {
			return fmt.Errorf("write main.js.min: %w", err)
		}
	}
	return nil
}
