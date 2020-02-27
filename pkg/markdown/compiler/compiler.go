package compiler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jschaf/b2/pkg/files"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/paths"
)

type Compiler struct {
	md *markdown.Markdown
}

func New(md *markdown.Markdown) *Compiler {
	return &Compiler{md}
}

// CompileAST compiles an AST into a writer.
func (c *Compiler) CompileAST(ast *markdown.PostAST, w io.Writer) error {
	b := &bytes.Buffer{}
	if err := c.md.Render(b, c.md.Source, ast); err != nil {
		return fmt.Errorf("failed to render markdown: %w", err)
	}

	data := html.TemplateData{
		Title: ast.Meta.Title,
		Body:  template.HTML(b.String()),
	}

	if err := html.PostDoc.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// CompileIntoDir compiles markdown into a directory based on the slug.
func (c *Compiler) CompileIntoDir(path string, r io.Reader, publicDir string) error {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read all file: %w", err)
	}

	postAST, err := c.md.Parse(path, bytes.NewReader(src))
	if err != nil {
		return fmt.Errorf("failed to parse markdown: %w", err)
	}

	slug := postAST.Meta.Slug
	if slug == "" {
		return fmt.Errorf("empty slug for path")
	}

	slugDir := filepath.Join(publicDir, slug)
	if err = os.MkdirAll(slugDir, 0755); err != nil {
		return fmt.Errorf("failed to make dir for slug %s: %w", slug, err)
	}

	dest := filepath.Join(slugDir, "index.html")
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open index.html file for write: %w", err)
	}

	if err := c.CompileAST(postAST, destFile); err != nil {
		return fmt.Errorf("failed to compile AST: %w", err)
	}

	for destPath, srcPath := range postAST.Assets {
		dest := filepath.Join(publicDir, destPath)
		if isSame, err := files.SameBytes(srcPath, dest); err != nil {
			return fmt.Errorf("failed to check if file contents are same: %w", err)
		} else if isSame {
			continue
		}

		if err := paths.Copy(srcPath, dest); err != nil {
			return fmt.Errorf("failed to copy asset to dest: %w", err)
		}
	}

	return nil
}

func CompileEverything(c *Compiler) error {
	rootDir, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("failed to find root git dir: %w", err)
	}
	postsDir := filepath.Join(rootDir, "posts")
	publicDir := filepath.Join(rootDir, "public")

	err = filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		return c.CompileIntoDir(file.Name(), file, publicDir)
	})

	if err != nil {
		return fmt.Errorf("failed to render markdown to HTML: %w", err)
	}
	return nil
}
