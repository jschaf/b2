package main

import (
	"fmt"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown/parser"
	"github.com/yuin/goldmark/renderer"
	"log"
	"os"
	"path/filepath"
)

func main() {
	rootDir, err := git.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root git dir: %s", err)
	}
	postsDir := filepath.Join(rootDir, "posts")
	publicDir := filepath.Join(rootDir, "public")
	md := parser.New()

	err = filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		node, err := md.Parse(file)
		if err != nil {
			return fmt.Errorf("failed to parse markdown: %w", err)
		}

		html, err := renderer.Render(bs)
		if err != nil {
			return err
		}

		os.MkdirAll(filepath.Join(publicDir))
	})

}
