package compiler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/markdown/mdext"
)

type IndexCompiler struct {
	md *markdown.Markdown
}

func NewForIndex(md *markdown.Markdown) *IndexCompiler {
	return &IndexCompiler{md: md}
}

func (ic *IndexCompiler) CompileASTs(asts []*markdown.PostAST, w io.Writer) error {
	bodies := make([]template.HTML, 0, len(asts))

	sort.Slice(asts, func(i, j int) bool {
		return asts[i].Meta.Date.After(asts[j].Meta.Date)
	})

	for _, ast := range asts {
		if ast.Meta.Visibility != mdext.VisibilityPublished {
			continue
		}

		b := new(bytes.Buffer)
		if err := ic.md.Render(b, ast.Source, ast); err != nil {
			return fmt.Errorf("failed to markdown for index: %w", err)
		}
		bodies = append(bodies, template.HTML(b.String()))
	}

	data := html.IndexTemplateData{
		Title:  "Joe Schafer's Blog",
		Bodies: bodies,
	}

	if err := html.IndexTemplate.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute index template: %w", err)
	}

	return nil
}

func (ic *IndexCompiler) CompileIntoDir(paths []string, rs []io.Reader, publicDir string) error {
	asts := make([]*markdown.PostAST, len(rs))
	for i, r := range rs {
		src, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read article for index: %w", err)
		}
		postAST, err := ic.md.Parse(paths[i], bytes.NewReader(src))
		if err != nil {
			return fmt.Errorf("failed to parse markdown for index: %w", err)
		}
		asts[i] = postAST
	}

	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return fmt.Errorf("failed to make dir for index: %w", err)
	}
	dest := filepath.Join(publicDir, "index.html")
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open index.html file for write: %w", err)
	}
	if err := ic.CompileASTs(asts, destFile); err != nil {
		return fmt.Errorf("failed to compile asts for index: %w", err)
	}

	return nil
}

func (ic *IndexCompiler) Compile() error {
	rootDir, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("failed to find root git dir: %w", err)
	}
	publicDir := filepath.Join(rootDir, "public")
	postsDir := filepath.Join(rootDir, "posts")

	paths := make([]string, 0, 16)
	readers := make([]io.Reader, 0, 16)
	err = filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		paths = append(paths, path)
		readers = append(readers, file)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse articles for index page: %w", err)
	}

	if err := ic.CompileIntoDir(paths, readers, publicDir); err != nil {
		return fmt.Errorf("failed to render index.html: %w", err)
	}
	return nil
}
