package html

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/jschaf/jsc/pkg/dirs"
	"github.com/jschaf/jsc/pkg/markdown/mdctx"

	"github.com/jschaf/jsc/pkg/git"
)

var (
	layoutDir = filepath.Join(git.RootDir(), dirs.Pkg, "markdown", "html")
	baseTmpl  = filepath.Join(layoutDir, "base.gohtml")

	indexTmpl = sync.OnceValue(func() *template.Template {
		tmpl := filepath.Join(layoutDir, "index.gohtml")
		return template.Must(
			template.New("index").Funcs(TemplateFuncs()).ParseFiles(tmpl, baseTmpl))
	})

	detailTmpl = sync.OnceValue(func() *template.Template {
		tmpl := filepath.Join(layoutDir, "detail.gohtml")
		return template.Must(
			template.New("detail").Funcs(TemplateFuncs()).ParseFiles(tmpl, baseTmpl))
	})
)

type IndexParams struct {
	Title    string
	Features *mdctx.FeatureSet
	Posts    []IndexPostParams
}

type IndexPostParams struct {
	Title     string
	TitleHTML template.HTML
	Slug      string
	Body      template.HTML
	Date      time.Time
}

func RenderIndex(w io.Writer, p IndexParams) error {
	err := indexTmpl().ExecuteTemplate(w, "base", p)
	if err != nil {
		return fmt.Errorf("execute index template: %w", err)
	}
	return nil
}

type DetailParams struct {
	Title    string
	Features *mdctx.FeatureSet
	Content  template.HTML
}

func RenderDetail(w io.Writer, p DetailParams) error {
	err := detailTmpl().ExecuteTemplate(w, "base", p)
	if err != nil {
		return fmt.Errorf("execute detail template: %w", err)
	}
	return nil
}
