package compiler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/jschaf/b2/pkg/paths"
	"github.com/karrick/godirwalk"
)

// DetailCompiler compiles the /* paths, showing the detail page for each
// post. Posts don't have another directory prefix.
type DetailCompiler struct {
	md     *markdown.Markdown
	pubDir string
}

// NewDetailCompiler creates a compiler for a post detail page.
func NewDetailCompiler(pubDir string) *DetailCompiler {
	md := markdown.New(markdown.WithHeadingAnchorStyle(mdext.HeadingAnchorStyleShow), markdown.WithTOCStyle(mdext.TOCStyleShow), markdown.WithExtender(mdext.NewNopContinueReadingExt()))
	return &DetailCompiler{md: md, pubDir: pubDir}
}

// parseFile parses a single post path into a markdown AST.
func (c *DetailCompiler) parseFile(path string) (*markdown.AST, error) {
	slog.Debug("compiling post detail", "path", path)
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
		return nil, fmt.Errorf("parseFile post markdown at path %s: %w", path, err)
	}
	return ast, nil
}

func (c *DetailCompiler) createDestFile(ast *markdown.AST) (*os.File, error) {
	slug := ast.Meta.Slug
	if slug == "" {
		return nil, fmt.Errorf("empty slug for path: %s", ast.Path)
	}
	slugDir := filepath.Join(c.pubDir, slug)
	if err := os.MkdirAll(slugDir, 0o755); err != nil {
		return nil, fmt.Errorf("make dir for slug %s: %w", slug, err)
	}
	dest := filepath.Join(slugDir, "index.html")
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create dest file %q for post: %w", dest, err)
	}
	return destFile, nil
}

// compileAST compiles a markdown AST into a writer.
func (c *DetailCompiler) compileAST(ast *markdown.AST, w io.Writer) error {
	b := &bytes.Buffer{}
	if err := c.md.Render(b, ast.Source, ast); err != nil {
		return fmt.Errorf("failed to render markdown: %w", err)
	}
	data := html.PostDetailData{
		Title:    ast.Meta.Title,
		Content:  template.HTML(b.String()),
		Features: ast.Features,
	}
	if err := html.RenderPostDetail(w, data); err != nil {
		return fmt.Errorf("failed to execute post template: %w", err)
	}

	if err := assets.CopyAll(c.pubDir, ast.Assets); err != nil {
		return err
	}
	return nil
}

func (c *DetailCompiler) Compile(glob string) error {
	err := c.compileDir(filepath.Join(git.RootDir(), dirs.Posts), glob)
	if err != nil {
		return fmt.Errorf("compile posts dir: %w", err)
	}
	err = c.compileDir(filepath.Join(git.RootDir(), dirs.TIL), glob)
	if err != nil {
		return fmt.Errorf("compile til dir: %w", err)
	}
	return nil
}

func (c *DetailCompiler) compileDir(dir string, glob string) (mErr error) {
	err := paths.WalkConcurrent(dir, runtime.NumCPU(), func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
		if glob != "" && !strings.Contains(path, glob) {
			return nil
		}

		ast, err := c.parseFile(path)
		if err != nil {
			return fmt.Errorf("parseFile TIL post into AST at path %s: %w", path, err)
		}

		dest, err := c.createDestFile(ast)
		if err != nil {
			return err
		}
		defer errs.Capture(&mErr, dest.Close, "close dest file")

		if err := c.compileAST(ast, dest); err != nil {
			return fmt.Errorf("compileAST AST for path %s: %w", path, err)
		}
		return nil
	})
	return err
}
