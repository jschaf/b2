package compiler

import (
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/jschaf/b2/pkg/paths"
	"github.com/karrick/godirwalk"
	"go.uber.org/zap"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const bookPostPrefix = "book"

// BookDetailCompiler compiles the /book/* paths, showing the detail pages for
// each book review.
type BookDetailCompiler struct {
	md     *markdown.Markdown
	pubDir string
	l      *zap.SugaredLogger
}

func NewBookDetail(pubDir string, l *zap.Logger) *BookDetailCompiler {
	md := markdown.New(l,
		markdown.WithHeadingAnchorStyle(mdext.HeadingAnchorStyleShow),
		markdown.WithTOCStyle(mdext.TOCStyleShow),
		markdown.WithExtender(mdext.NewNopContinueReadingExt()),
	)
	return &BookDetailCompiler{md: md, pubDir: pubDir, l: l.Sugar()}
}

func (c *BookDetailCompiler) parse(path string) (*markdown.AST, error) {
	c.l.Debugf("compiling book detail %s", path)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open TIL post %s: %w", path, err)
	}
	src, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read post at path %s: %w", path, err)
	}
	ast, err := c.md.Parse(path, bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("parse post markdown at path %s: %w", path, err)
	}
	return ast, nil
}

func (c *BookDetailCompiler) createDestFile(ast *markdown.AST) (*os.File, error) {
	slug := ast.Meta.Slug
	if slug == "" {
		return nil, fmt.Errorf("empty slug for path: %s", ast.Path)
	}
	slugDir := filepath.Join(c.pubDir, bookPostPrefix, slug)
	if err := os.MkdirAll(slugDir, 0755); err != nil {
		return nil, fmt.Errorf("make dir for slug %s: %w", slug, err)
	}
	dest := filepath.Join(slugDir, "index.html")
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("create dest file %q for book detail: %w", dest, err)
	}
	return destFile, nil
}

func (c *BookDetailCompiler) compile(ast *markdown.AST, w io.Writer) error {
	b := &bytes.Buffer{}
	if err := c.md.Render(b, ast.Source, ast); err != nil {
		return fmt.Errorf("failed to render markdown: %w", err)
	}
	data := html.BookDetailData{
		Title:    ast.Meta.Title,
		Content:  template.HTML(b.String()),
		Features: ast.Features,
	}
	if err := html.RenderBookDetail(w, data); err != nil {
		return fmt.Errorf("failed to execute post template: %w", err)
	}

	if err := assets.CopyAll(c.pubDir, ast.Assets); err != nil {
		return err
	}
	return nil
}

func (c *BookDetailCompiler) CompileAll() error {
	postsDir := filepath.Join(git.MustFindRootDir(), dirs.Book)
	err := paths.WalkConcurrent(postsDir, runtime.NumCPU(), func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
		ast, err := c.parse(path)
		if err != nil {
			return fmt.Errorf("parse TIL post into AST at path %s: %w", path, err)
		}
		dest, err := c.createDestFile(ast)
		if err != nil {
			return err
		}
		if err := c.compile(ast, dest); err != nil {
			return fmt.Errorf("compile AST for path %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("compile all posts: %w", err)
	}
	return nil
}
