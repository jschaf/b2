package main

import (
	"github.com/jschaf/b2/pkg/dirs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/js"
)

func main() {
	rootDir := git.MustFindRootDir()
	pubDir := filepath.Join(rootDir, dirs.Public)
	result, err := js.BundleMain(pubDir)
	if err != nil {
		log.Fatal(err)
	}
	out := filepath.Join(rootDir, "dist", "main.min.js")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(result.JsAbsPath, result.JsContents, 0644); err != nil {
		log.Fatal(err)
	}
	if err = ioutil.WriteFile(result.SourceMapAbsPath, result.SourceMapContents, 0644); err != nil {
		log.Fatal(err)
	}
}
