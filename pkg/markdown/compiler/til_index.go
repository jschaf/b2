package compiler

import (
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/jschaf/b2/pkg/paths"
	"github.com/karrick/godirwalk"
	"go.uber.org/zap"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// TILIndexCompiler compiles the /til/ path, an index of all TIL posts.
type TILIndexCompiler struct {
	md     *markdown.Markdown
	l      *zap.SugaredLogger
	pubDir string
}

func NewTILIndex(pubDir string, l *zap.Logger) *TILIndexCompiler {
	md := markdown.New(l)
	return &TILIndexCompiler{md: md, pubDir: pubDir, l: l.Sugar()}
}

func (c *TILIndexCompiler) parse(path string) (*markdown.AST, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open TIL post %s: %w", path, err)
	}
	src, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read TIL filepath %s: %w", path, err)
	}
	ast, err := c.md.Parse(path, bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("parse TIL markdown at path %s: %w", path, err)
	}
	return ast, nil
}

func (c *TILIndexCompiler) compileASTs(asts []*markdown.AST, w io.Writer) error {
	bodies := make([]template.HTML, 0, len(asts))
	sort.Slice(asts, func(i, j int) bool {
		return asts[i].Meta.Date.After(asts[j].Meta.Date)
	})
	feats := mdctx.NewFeatures()
	for _, ast := range asts {
		if ast.Meta.Visibility != mdext.VisibilityPublished {
			continue
		}
		b := new(bytes.Buffer)
		if err := c.md.Render(b, ast.Source, ast); err != nil {
			return fmt.Errorf("failed to markdown for index: %w", err)
		}
		bodies = append(bodies, template.HTML(b.String()))
		feats.AddAll(ast.Features)
	}
	data := html.TILIndexData{
		Title:    "TIL - Joe Schafer's Blog",
		Bodies:   bodies,
		Features: feats,
	}
	if err := html.RenderTILIndex(w, data); err != nil {
		return fmt.Errorf("render TIL: %w", err)
	}
	return nil
}

func (c *TILIndexCompiler) CompileIndex() error {
	tilDir := filepath.Join(git.MustFindRootDir(), dirs.TIL)

	astsC := make(chan *markdown.AST)
	asts := make([]*markdown.AST, 0, 16)

	done := make(chan struct{})
	go func() {
		for ast := range astsC {
			asts = append(asts, ast)
		}
		close(done)
	}()

	err := paths.WalkConcurrent(tilDir, runtime.NumCPU(), func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
		c.l.Debugf("compiling til index %s", path)
		ast, err := c.parse(path)
		if err != nil {
			return fmt.Errorf("parse TIL into AST for index at path %s: %w", path, err)
		}
		astsC <- ast
		return nil
	})
	if err != nil {
		return fmt.Errorf("compile all TILs: %w", err)
	}

	close(astsC)
	<-done

	dest := filepath.Join(c.pubDir, dirs.TIL, "index.html")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("make dir for TILs: %w", err)
	}
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open TIL index.html for write: %w", err)
	}
	if err := c.compileASTs(asts, destFile); err != nil {
		return fmt.Errorf("compile asts for TILs: %w", err)
	}
	return nil
}
