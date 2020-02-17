package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
)

func main() {
	rootDir, err := git.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root git dir: %s", err)
	}
	postsDir := filepath.Join(rootDir, "posts")
	publicDir := filepath.Join(rootDir, "public")
	md := markdown.New()
	c := compiler.New(md)

	err = filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		log.Printf("rendering file %s", path)

		return c.CompileIntoDir(file, md, publicDir)
	})

	if err != nil {
		log.Printf("failed to render markdown to HTML: %s", err)
	}

}
