package compiler

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/jschaf/b2/pkg/paths"
)

// RootIndexCompiler compiles the / path, the main homepage.
type RootIndexCompiler struct {
	md     *markdown.Markdown
	pubDir string
}

func NewRootIndex(pubDir string) *RootIndexCompiler {
	md := markdown.New(markdown.WithExtender(mdext.NewContinueReadingExt()))
	return &RootIndexCompiler{md: md, pubDir: pubDir}
}

func (ic *RootIndexCompiler) parsePosts() ([]*markdown.AST, error) {
	postsDir := filepath.Join(git.RootDir(), dirs.Posts)
	asts, err := paths.WalkCollect(postsDir, func(path string, dirent fs.DirEntry) ([]*markdown.AST, error) {
		if !dirent.Type().IsRegular() || filepath.Ext(path) != ".md" {
			return nil, nil
		}
		slog.Debug("compiling for index", "path", path)
		bs, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read post at path %s: %w", path, err)
		}
		ast, err := ic.md.Parse(path, bytes.NewReader(bs))
		if err != nil {
			return nil, fmt.Errorf("parse markdown for root index: %w", err)
		}
		return []*markdown.AST{ast}, nil
	})
	return asts, err
}

func (ic *RootIndexCompiler) renderPosts(asts []*markdown.AST) ([]html.RootPostData, error) {
	posts := make([]html.RootPostData, 0, len(asts))
	for _, ast := range asts {
		if ast.Meta.Visibility != mdext.VisibilityPublished {
			continue
		}
		b := new(bytes.Buffer)
		if err := ic.md.Render(b, ast.Source, ast); err != nil {
			return nil, fmt.Errorf("render markdown for index: %w", err)
		}
		posts = append(posts, html.RootPostData{
			Title: ast.Meta.Title,
			Slug:  ast.Meta.Slug,
			Date:  ast.Meta.Date,
			Body:  template.HTML(b.String()),
		})
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })
	return posts, nil
}

func (ic *RootIndexCompiler) CompileIndex() error {
	postASTs, err := ic.parsePosts()
	if err != nil {
		return err
	}

	featureSet := mdctx.NewFeatureSet()
	for _, ast := range postASTs {
		featureSet.AddAll(ast.Features)
	}

	posts, err := ic.renderPosts(postASTs)
	if err != nil {
		return fmt.Errorf("compile postASTs for index: %w", err)
	}

	if err := os.MkdirAll(ic.pubDir, 0o755); err != nil {
		return fmt.Errorf("make dir for index: %w", err)
	}
	dest := filepath.Join(ic.pubDir, "index.html")

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open index.html file for write: %w", err)
	}
	data := html.RootIndexData{
		Title:    "Joe Schafer's Blog",
		Posts:    posts,
		Features: featureSet,
	}
	if err := html.RenderRootIndex(destFile, data); err != nil {
		return fmt.Errorf("execute index template: %w", err)
	}

	return nil
}
