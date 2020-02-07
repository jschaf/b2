package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/html"
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

		src, err := ioutil.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read all file: %w", err)
		}

		postAST, err := md.Parse(bytes.NewReader(src))
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

		r := &bytes.Buffer{}
		if err = md.Render(r, src, postAST); err != nil {
			return fmt.Errorf("failed to render markdown: %w", err)
		}

		data := html.TemplateData{
			Title: postAST.Meta.Title,
			Body:  template.HTML(r.String()),
		}
		if err = html.PostDoc.Execute(destFile, data); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	})

}
