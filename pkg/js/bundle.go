package js

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/esbuild/pkg/bundler"
	"github.com/jschaf/esbuild/pkg/fs"
	"github.com/jschaf/esbuild/pkg/logging"
	"github.com/jschaf/esbuild/pkg/parser"
	"github.com/jschaf/esbuild/pkg/resolver"
)

type jsCache struct {
	mu           sync.Mutex
	key          uint64
	bundleResult bundler.BundleResult
}

// Single entry cache for main.js.
var mainJSCache = jsCache{}

// BundleMain minifies the main.js file and returns the bytes of the written file.
func BundleMain() (bundler.BundleResult, error) {
	rootDir, err := git.FindRootDir()
	if err != nil {
		return bundler.BundleResult{}, fmt.Errorf("failed to find root git dir: %w", err)
	}
	outDir := filepath.Join(rootDir, "public")

	logOpts := logging.StderrOptions{
		IncludeSource:      true,
		ErrorLimit:         3,
		ExitWhenLimitIsHit: true,
	}
	stderrLog, join := logging.NewStderrLog(logOpts)
	realFS := fs.RealFS()
	fsResolver := resolver.NewResolver(realFS, []string{".js"})
	parseOpts := parser.ParseOptions{
		IsBundling:           true,
		Defines:              nil,
		MangleSyntax:         true,
		KeepSingleExpression: false,
		OmitWarnings:         false,
		JSX:                  parser.JSXOptions{},
		Target:               parser.ES2019,
	}
	bundleOpts := bundler.BundleOptions{
		Bundle:            false,
		AbsOutputFile:     "",
		AbsOutputDir:      outDir,
		RemoveWhitespace:  true,
		MinifyIdentifiers: true,
		MangleSyntax:      true,
		SourceMap:         true,
		ModuleName:        "",
		ExtensionToLoader: nil,
	}
	entryPoint := filepath.Join(rootDir, "scripts", "main.js")
	var curKey, newKey uint64
	if bytes, err := ioutil.ReadFile(entryPoint); err != nil {
		return bundler.BundleResult{}, fmt.Errorf("failed to read entrypoint: %w", err)
	} else {
		mainJSCache.mu.Lock()
		curKey = mainJSCache.key
		curResult := mainJSCache.bundleResult
		mainJSCache.mu.Unlock()

		hasher := fnv.New64a()
		_, _ = hasher.Write(bytes)
		newKey = hasher.Sum64()
		if newKey == curKey {
			return curResult, nil
		}
	}
	bundle := bundler.ScanBundle(
		stderrLog, realFS, fsResolver, []string{entryPoint}, parseOpts, bundleOpts)
	if join().Errors != 0 {
		return bundler.BundleResult{}, fmt.Errorf("bundleResult scanning had errors: %s", join())
	}

	log2, join2 := logging.NewStderrLog(logOpts)
	result := bundle.Compile(log2, bundleOpts)

	// Early exit if there were errors.
	if join2().Errors != 0 {
		return bundler.BundleResult{}, fmt.Errorf("bundleResult compilation had errors: %s", join())
	}

	// Return the first result, ignore other entry points.
	if len(result) > 1 {
		return bundler.BundleResult{}, errors.New("got more than 1 result")
	}

	item := result[0]
	mainJSCache.mu.Lock()
	mainJSCache.key = newKey
	mainJSCache.bundleResult = item
	mainJSCache.mu.Unlock()
	return item, nil
}

func WriteMainBundle(result bundler.BundleResult) error {
	rootDir, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("bundle main.js find git root: %w", err)
	}
	out := filepath.Join(rootDir, "public", "main.min.js")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return fmt.Errorf("mkdir -p for main js bundles: %w", err)
	}

	if err = ioutil.WriteFile(result.JsAbsPath, result.JsContents, 0644); err != nil {
		return fmt.Errorf("write main.js.min: %w", err)
	}
	if err = ioutil.WriteFile(result.SourceMapAbsPath, result.SourceMapContents, 0644); err != nil {
		return fmt.Errorf("write error for main.js source map: %w", err)
	}
	return nil
}
