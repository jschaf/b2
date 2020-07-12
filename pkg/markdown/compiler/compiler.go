package compiler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"go.uber.org/zap"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jschaf/b2/pkg/css"
	"github.com/jschaf/b2/pkg/files"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/paths"
)

type Compiler struct {
	md     *markdown.Markdown
	logger *zap.SugaredLogger
}

// NewForPostDetail creates a compiler for a post detail page.
func NewForPostDetail(l *zap.Logger) *Compiler {
	md := markdown.New(l,
		markdown.WithHeadingAnchorStyle(mdext.HeadingAnchorStyleShow),
		markdown.WithTOCStyle(mdext.TOCStyleShow),
		markdown.WithExtender(mdext.NewNopContinueReadingExt()),
	)
	return &Compiler{md: md, logger: l.Sugar()}
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

func CleanPubDir(rootDir string) error {
	publicDir := filepath.Join(rootDir, "public")

	if stat, err := os.Stat(publicDir); err == nil {
		if !stat.IsDir() {
			return errors.New("public dir is not a directory")
		}
		if err := os.RemoveAll(publicDir); err != nil {
			return fmt.Errorf("failed to delete public dir: %w", err)
		}
	} else if os.IsNotExist(err) {
		// Do nothing.
	} else {
		return fmt.Errorf("failed to stat pub dir: %w", err)
	}

	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return fmt.Errorf("failed to make public dir: %w", err)
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
	rootDir, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("failed to find root git dir: %w", err)
	}
	postsDir := filepath.Join(rootDir, "posts")
	publicDir := filepath.Join(rootDir, "public")
	if err := CleanPubDir(rootDir); err != nil {
		return fmt.Errorf("failed to clean public dir: %w", err)
	}

	err = filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" || strings.HasSuffix(path, ".previews.md") {
			return nil
		}
		if glob != "" && !strings.Contains(path, glob) {
			return nil
		}

		c.logger.Debugf("compiling %s", path)
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		return c.CompileIntoDir(file.Name(), file, publicDir)
	})

	if _, err := css.WriteMainCSS(rootDir); err != nil {
		return fmt.Errorf("failed to compile main.css: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to render markdown to HTML: %w", err)
	}

	return nil
}
