package main

import (
	"github.com/jschaf/b2/pkg/dirs"
	"log"
	"path/filepath"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/js"
)

func main() {
	rootDir := git.MustFindRootDir()
	pubDir := filepath.Join(rootDir, dirs.Public)
	if err := js.WriteMainBundle(pubDir); err != nil {
		log.Fatal(err)
	}
}
