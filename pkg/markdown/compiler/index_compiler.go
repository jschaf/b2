package compiler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/karrick/godirwalk"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

type IndexCompiler struct {
	md *markdown.Markdown
	l  *zap.SugaredLogger
}

func NewForIndex(l *zap.Logger) *IndexCompiler {
	md := markdown.New(l, markdown.WithExtender(mdext.NewContinueReadingExt()))
	return &IndexCompiler{md: md, l: l.Sugar()}
}

func (ic *IndexCompiler) compileASTs(asts []*markdown.PostAST, w io.Writer) error {
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

	if err := html.RenderIndex(w, data); err != nil {
		return fmt.Errorf("failed to execute index template: %w", err)
	}

	return nil
}

func (ic *IndexCompiler) compilePost(path string, r io.Reader) (*markdown.PostAST, error) {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read article for index: %w", err)
	}
	postAST, err := ic.md.Parse(path, bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown for index: %w", err)
	}
	return postAST, nil
}

func (ic *IndexCompiler) Compile() error {
	rootDir := git.MustFindRootDir()
	publicDir := filepath.Join(rootDir, "public")
	postsDir := filepath.Join(rootDir, "posts")

	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	g, ctx := errgroup.WithContext(context.Background())
	astsC := make(chan *markdown.PostAST)
	asts := make([]*markdown.PostAST, 0, 16)

	walkFunc := func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
		if err := sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("acquire semaphore: %w", err)
		}

		g.Go(func() error {
			defer sem.Release(1)
			ic.l.Debugf("compiling for index %s", path)
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("open file %s: %w", path, err)
			}
			ast, err := ic.compilePost(path, file)
			if err != nil {
				return fmt.Errorf("compile into ast %s: %w", path, err)
			}
			astsC <- ast
			return nil
		})
		return nil
	}

	err := godirwalk.Walk(
		postsDir, &godirwalk.Options{Unsorted: true, Callback: walkFunc})
	if err != nil {
		return fmt.Errorf("index compiler walk: %w", err)
	}

	done := make(chan struct{})
	go func() {
		for ast := range astsC {
			asts = append(asts, ast)
		}
		close(done)
	}()

	if err := g.Wait(); err != nil {
		return fmt.Errorf("index compiler wait err group: %w", err)
	}
	close(astsC)
	<-done

	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return fmt.Errorf("failed to make dir for index: %w", err)
	}
	dest := filepath.Join(publicDir, "index.html")
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open index.html file for write: %w", err)
	}
	if err := ic.compileASTs(asts, destFile); err != nil {
		return fmt.Errorf("failed to compile asts for index: %w", err)
	}

	return nil
}
