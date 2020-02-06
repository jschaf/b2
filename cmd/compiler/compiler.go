package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
)

func main() {
	rootDir, err := git.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root git dir: %s", err)
	}
	postsDir := filepath.Join(rootDir, "posts")
	publicDir := filepath.Join(rootDir, "public")
	md := markdown.New()

	err = filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		log.Printf("rendering file %s", path)

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		bs, err := ioutil.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read all file: %w", err)
		}

		postAST, err := md.Parse(bytes.NewReader(bs))
		if err != nil {
			return fmt.Errorf("failed to parse markdown: %w", err)
		}

		slug := postAST.Meta.Slug
		if slug == "" {
			return fmt.Errorf("empty slug for path: %s", path)
		}

		slugDir := filepath.Join(publicDir, slug)
		if err = os.MkdirAll(slugDir, 0666); err != nil {
			return fmt.Errorf("failed to make dir for slug %s: %w", slug, err)
		}

		dest := filepath.Join(slugDir, "index.html")
		destFile, err := os.OpenFile(dest, os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("failed to open index.html file for write: %w", err)
		}

		err = md.Render(destFile, bs, postAST)
		if err != nil {
			return fmt.Errorf("failed to render file: %w", err)
		}

		return nil
	})

}
