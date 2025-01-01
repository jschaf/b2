package js

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	esbuild "github.com/evanw/esbuild/pkg/api"
	"github.com/jschaf/jsc/pkg/dirs"
	"github.com/jschaf/jsc/pkg/files"
	"github.com/jschaf/jsc/pkg/git"
)

// Single entry cache for main.js.
type jsCache struct {
	mu           sync.Mutex
	key          uint64
	bundleResult esbuild.BuildResult
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

func bundleTypeScript(distDir string) (esbuild.BuildResult, error) {
	mainTS := filepath.Join(git.RootDir(), dirs.Static, "main.ts")
	mainTSOut := filepath.Join(distDir, "main.js")

	// Check if the file is the same; if so, skip the bundle step.
	mainJSCache.mu.Lock()
	defer mainJSCache.mu.Unlock()
	ok, newKey, err := mainJSCache.isUnchanged(mainTS)
	if err != nil {
		return esbuild.BuildResult{}, err
	} else if ok {
		return mainJSCache.bundleResult, nil
	}

	result := esbuild.Build(esbuild.BuildOptions{
		EntryPoints: []string{mainTS},
		Outfile:     mainTSOut,
		Target:      esbuild.ES2019,
		Bundle:      false,
		Write:       true,
		LogLevel:    esbuild.LogLevelError,
	})
	if len(result.Errors) > 0 {
		msg := ""
		for _, e := range result.Errors {
			msg += "\n" + e.Text
		}
		return esbuild.BuildResult{}, fmt.Errorf("bundle typescript errors:" + msg)
	}

	mainJSCache.key = newKey
	mainJSCache.bundleResult = result
	return result, nil
}

func WriteTypeScriptMain(distDir string) error {
	result, err := bundleTypeScript(distDir)
	if err != nil {
		return fmt.Errorf("write main.ts bundle: %w", err)
	}
	for _, f := range result.OutputFiles {
		if err := os.WriteFile(f.Path, f.Contents, 0o644); err != nil {
			return fmt.Errorf("write main.js.min: %w", err)
		}
	}
	return nil
}
