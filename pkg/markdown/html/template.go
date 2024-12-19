package html

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"time"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/markdown/mdctx"

	"github.com/jschaf/b2/pkg/git"
)

func compileTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template, 8)
	rootDir := git.RootDir()
	layoutDir := filepath.Join(rootDir, dirs.Pkg, "markdown", "html")
	baseTmpl := filepath.Join(layoutDir, "base.gohtml")
	layouts := []string{
		"detail.gohtml",
		"index.gohtml",
	}
	for _, name := range layouts {
		f := filepath.Join(layoutDir, name)
		templates[name] = template.Must(
			template.New(name).Funcs(TemplateFuncs()).ParseFiles(f, baseTmpl))
	}
	return templates
}

func render(w io.Writer, name string, data map[string]any) error {
	templates := compileTemplates()
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("template %s does not exist", name)
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}

type PostDetailData struct {
	Title    string
	Features *mdctx.FeatureSet
	Content  template.HTML
}

func RenderPostDetail(w io.Writer, d PostDetailData) error {
	m := map[string]any{
		"Title":      d.Title,
		"Content":    d.Content,
		"FeatureSet": d.Features,
	}
	return render(w, "detail.gohtml", m)
}

type RootPostData struct {
	Title string
	Slug  string
	Body  template.HTML
	Date  time.Time
}

type RootIndexData struct {
	Title    string
	Features *mdctx.FeatureSet
	Posts    []RootPostData
}

func RenderRootIndex(w io.Writer, d RootIndexData) error {
	m := map[string]any{
		"Title":      d.Title,
		"Posts":      d.Posts,
		"FeatureSet": d.Features,
	}
	return render(w, "index.gohtml", m)
}
