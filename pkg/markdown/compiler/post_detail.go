package compiler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/karrick/godirwalk"
	"go.uber.org/zap"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/paths"
)

// PostDetailCompiler compiles the /* paths, showing the detail page for each
// post. Posts don't have another directory prefix.
type PostDetailCompiler struct {
	md     *markdown.Markdown
	pubDir string
	l      *zap.SugaredLogger
}

// NewPostDetail creates a compiler for a post detail page.
func NewPostDetail(pubDir string, l *zap.Logger) *PostDetailCompiler {
	md := markdown.New(l,
		markdown.WithHeadingAnchorStyle(mdext.HeadingAnchorStyleShow),
		markdown.WithTOCStyle(mdext.TOCStyleShow),
		markdown.WithExtender(mdext.NewNopContinueReadingExt()),
	)
	return &PostDetailCompiler{md: md, pubDir: pubDir, l: l.Sugar()}
}

// parse parses a single post path into a markdown AST.
func (c *PostDetailCompiler) parse(path string) (*markdown.AST, error) {
	c.l.Debugf("compiling post detail %s", path)
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

func (c *PostDetailCompiler) createDestFile(ast *markdown.AST) (*os.File, error) {
	slug := ast.Meta.Slug
	if slug == "" {
		return nil, fmt.Errorf("empty slug for path: %s", ast.Path)
	}
	slugDir := filepath.Join(c.pubDir, slug)
	if err := os.MkdirAll(slugDir, 0755); err != nil {
		return nil, fmt.Errorf("make dir for slug %s: %w", slug, err)
	}
	dest := filepath.Join(slugDir, "index.html")
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("create dest file %q for post: %w", dest, err)
	}
	return destFile, nil
}

// compile compiles a markdown AST into a writer.
func (c *PostDetailCompiler) compile(ast *markdown.AST, w io.Writer) error {
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

func (c *PostDetailCompiler) CompileAll(glob string) error {
	postsDir := filepath.Join(git.RootDir(), dirs.Posts)
	err := paths.WalkConcurrent(postsDir, runtime.NumCPU(), func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
		if glob != "" && !strings.Contains(path, glob) {
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
