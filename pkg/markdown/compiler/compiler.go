package compiler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/karrick/godirwalk"
	"go.uber.org/zap"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jschaf/b2/pkg/css"
	"github.com/jschaf/b2/pkg/files"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/paths"
)

type Compiler struct {
	md *markdown.Markdown
	l  *zap.SugaredLogger
}

// NewForPostDetail creates a compiler for a post detail page.
func NewForPostDetail(l *zap.Logger) *Compiler {
	md := markdown.New(l,
		markdown.WithHeadingAnchorStyle(mdext.HeadingAnchorStyleShow),
		markdown.WithTOCStyle(mdext.TOCStyleShow),
		markdown.WithExtender(mdext.NewNopContinueReadingExt()),
	)
	return &Compiler{md: md, l: l.Sugar()}
}

// CompileAST compiles an AST into a writer.
func (c *Compiler) CompileAST(ast *markdown.PostAST, w io.Writer) error {
	b := &bytes.Buffer{}
	if err := c.md.Render(b, ast.Source, ast); err != nil {
		return fmt.Errorf("failed to render markdown: %w", err)
	}

	data := html.PostTemplateData{
		Title:   ast.Meta.Title,
		Content: template.HTML(b.String()),
	}
	if err := html.RenderPost(w, data); err != nil {
		return fmt.Errorf("failed to execute post template: %w", err)
	}

	return nil
}

// CompileIntoDir compiles markdown into the public directory based on the slug.
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
		if isSame, err := files.SameBytes(srcPath, dest); errors.Is(err, os.ErrNotExist) {
			// Ignore
		} else if err != nil {
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

func (c *Compiler) CompileAllPosts(glob string) error {
	rootDir := git.MustFindRootDir()
	postsDir := filepath.Join(rootDir, dirs.Posts)
	publicDir := filepath.Join(rootDir, dirs.Public)

	err := paths.WalkConcurrent(postsDir, runtime.NumCPU(), func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
		if glob != "" && !strings.Contains(path, glob) {
			return nil
		}

		c.l.Debugf("compiling %s", path)
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		err = c.CompileIntoDir(file.Name(), file, publicDir)
		if err != nil {
			return fmt.Errorf("compile post %q: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("compile walk: %w", err)
	}

	if _, err := css.WriteMainCSS(rootDir); err != nil {
		return fmt.Errorf("failed to compile main.css: %w", err)
	}

	return nil
}
