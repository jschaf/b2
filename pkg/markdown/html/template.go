package html

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"reflect"
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
		"book_detail.gohtml",
		"post_detail.gohtml",
		"root_index.gohtml",
		"til_detail.gohtml",
		"til_index.gohtml",
	}
	for _, name := range layouts {
		f := filepath.Join(layoutDir, name)
		templates[name] = template.Must(
			template.New(name).Funcs(TemplateFuncs()).ParseFiles(f, baseTmpl))
	}
	return templates
}

func render(w io.Writer, name string, data map[string]interface{}) error {
	templates := compileTemplates()
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("template %s does not exist", name)
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}

type BookDetailData struct {
	Title    string
	Features *mdctx.FeatureSet
	Content  template.HTML
}

func RenderBookDetail(w io.Writer, d BookDetailData) error {
	m := map[string]interface{}{
		"Title":      d.Title,
		"Content":    d.Content,
		"FeatureSet": d.Features,
	}
	return render(w, "book_detail.gohtml", m)
}

type PostDetailData struct {
	Title    string
	Features *mdctx.FeatureSet
	Content  template.HTML
}

func RenderPostDetail(w io.Writer, d PostDetailData) error {
	m := map[string]interface{}{
		"Title":      d.Title,
		"Content":    d.Content,
		"FeatureSet": d.Features,
	}
	return render(w, "post_detail.gohtml", m)
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
	TILs     []TILIndexData
}

func RenderRootIndex(w io.Writer, d RootIndexData) error {
	m := map[string]interface{}{
		"Title":      d.Title,
		"Posts":      d.Posts,
		"TILs":       d.TILs,
		"FeatureSet": d.Features,
	}
	return render(w, "root_index.gohtml", m)
}

type TILIndexData struct {
	Title    string
	Features *mdctx.FeatureSet
	Bodies   []template.HTML
}

func RenderTILIndex(w io.Writer, d TILIndexData) error {
	m := map[string]interface{}{
		"Title":      d.Title,
		"Bodies":     d.Bodies,
		"FeatureSet": d.Features,
	}
	return render(w, "til_index.gohtml", m)
}

type TILDetailData struct {
	Title    string
	Features *mdctx.FeatureSet
	Content  template.HTML
}

func RenderTILDetail(w io.Writer, d TILDetailData) error {
	m := make(map[string]interface{})
	m["Title"] = d.Title
	m["Content"] = d.Content
	m["FeatureSet"] = d.Features
	return render(w, "til_detail.gohtml", m)
}

// isLast returns true if index is the last index in item.
func isLast(index int, item interface{}) (bool, error) {
	v := reflect.ValueOf(item)
	if !v.IsValid() {
		return false, fmt.Errorf("isLast of untyped nil")
	}
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return index == v.Len()-1, nil
	}
	return false, fmt.Errorf("isLast of type %s", v.Type())
}
