package compiler

import (
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/html"
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

type IndexCompiler struct {
	md     *markdown.Markdown
	l      *zap.SugaredLogger
	pubDir string
}

func NewForIndex(pubDir string, l *zap.Logger) *IndexCompiler {
	md := markdown.New(l, markdown.WithExtender(mdext.NewContinueReadingExt()))
	return &IndexCompiler{md: md, pubDir: pubDir, l: l.Sugar()}
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

	if err := html.RenderIndex(ic.pubDir, w, data); err != nil {
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
	postsDir := filepath.Join(git.MustFindRootDir(), dirs.Posts)

	astsC := make(chan *markdown.PostAST)
	asts := make([]*markdown.PostAST, 0, 16)

	done := make(chan struct{})
	go func() {
		for ast := range astsC {
			asts = append(asts, ast)
		}
		close(done)
	}()

	err := paths.WalkConcurrent(postsDir, runtime.NumCPU(), func(path string, dirent *godirwalk.Dirent) error {
		if !dirent.IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
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
	if err != nil {
		return fmt.Errorf("index compiler walk: %w", err)
	}

	close(astsC)
	<-done

	if err := os.MkdirAll(ic.pubDir, 0755); err != nil {
		return fmt.Errorf("failed to make dir for index: %w", err)
	}
	dest := filepath.Join(ic.pubDir, "index.html")
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open index.html file for write: %w", err)
	}
	if err := ic.compileASTs(asts, destFile); err != nil {
		return fmt.Errorf("failed to compile asts for index: %w", err)
	}

	return nil
}
