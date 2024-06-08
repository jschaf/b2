package js

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/files"
	"github.com/jschaf/b2/pkg/git"
)

// Single entry cache for main.js.
type jsCache struct {
	mu           sync.Mutex
	key          uint64
	bundleResult api.BuildResult
}

var mainJSCache = &jsCache{}

func (jsCache *jsCache) isUnchanged(mainJSPath string) (bool, uint64, error) {
	newKey, err := files.HashContentsFnv64(mainJSPath)
	if err != nil {
		return false, 0, fmt.Errorf("hash main.js for JS build cache: %w", err)
	}
	curKey := mainJSCache.key
	if newKey == curKey {
		return true, newKey, nil
	}
	return false, newKey, nil
}

func bundleTypeScript(pubDir string) (api.BuildResult, error) {
	mainTS := filepath.Join(git.RootDir(), dirs.Static, "main.ts")
	mainTSOut := filepath.Join(pubDir, "main.js")

	// Check if file is same and skip JS bundle.
	mainJSCache.mu.Lock()
	defer mainJSCache.mu.Unlock()
	ok, newKey, err := mainJSCache.isUnchanged(mainTS)
	if err != nil {
		return api.BuildResult{}, err
	} else if ok {
		return mainJSCache.bundleResult, nil
	}

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{mainTS},
		Outfile:     mainTSOut,
		Target:      api.ES2019,
		Bundle:      false,
		Write:       true,
		Loaders: map[string]api.Loader{
			".ts": api.LoaderTS,
		},
		ErrorLimit: 3,
		LogLevel:   api.LogLevelInfo,
	})
	if len(result.Errors) > 0 {
		msg := ""
		for _, e := range result.Errors {
			msg += "\n" + e.Text
		}
		return api.BuildResult{}, fmt.Errorf("bundle typescript errors:" + msg)
	}

	mainJSCache.key = newKey
	mainJSCache.bundleResult = result
	return result, nil
}

func WriteTypeScriptMain(pubDir string) error {
	result, err := bundleTypeScript(pubDir)
	if err != nil {
		return fmt.Errorf("write main.ts bundle: %w", err)
	}
	for _, f := range result.OutputFiles {
		if err := ioutil.WriteFile(f.Path, f.Contents, 0o644); err != nil {
			return fmt.Errorf("write main.js.min: %w", err)
		}
	}
	return nil
}
