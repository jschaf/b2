package html

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"html/template"
	"io"
	"path/filepath"
	"reflect"

	"github.com/jschaf/b2/pkg/git"
)

var fns = template.FuncMap{
	"isLast": isLast,
}

var (
	templates = make(map[string]*template.Template)
)

func init() {
	rootDir := git.MustFindRootDir()
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
			template.New(name).Funcs(fns).ParseFiles(f, baseTmpl))
	}
}

func render(w io.Writer, name string, data map[string]interface{}) error {
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("template %s does not exist", name)
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}

type BookDetailData struct {
	Title    string
	Features *mdctx.Features
	Content  template.HTML
}

func RenderBookDetail(w io.Writer, d BookDetailData) error {
	m := map[string]interface{}{
		"Title":    d.Title,
		"Content":  d.Content,
		"Features": d.Features,
	}
	return render(w, "book_detail.gohtml", m)
}

type PostDetailData struct {
	Title    string
	Features *mdctx.Features
	Content  template.HTML
}

func RenderPostDetail(w io.Writer, d PostDetailData) error {
	m := map[string]interface{}{
		"Title":    d.Title,
		"Content":  d.Content,
		"Features": d.Features,
	}
	return render(w, "post_detail.gohtml", m)
}

type RootIndexData struct {
	Title    string
	Features *mdctx.Features
	Bodies   []template.HTML
}

func RenderRootIndex(w io.Writer, d RootIndexData) error {
	m := map[string]interface{}{
		"Title":    d.Title,
		"Bodies":   d.Bodies,
		"Features": d.Features,
	}
	return render(w, "root_index.gohtml", m)
}

type TILIndexData struct {
	Title    string
	Features *mdctx.Features
	Bodies   []template.HTML
}

func RenderTILIndex(w io.Writer, d TILIndexData) error {
	m := map[string]interface{}{
		"Title":    d.Title,
		"Bodies":   d.Bodies,
		"Features": d.Features,
	}
	return render(w, "til_index.gohtml", m)
}

type TILDetailData struct {
	Title    string
	Features *mdctx.Features
	Content  template.HTML
}

func RenderTILDetail(w io.Writer, d TILDetailData) error {
	m := make(map[string]interface{})
	m["Title"] = d.Title
	m["Content"] = d.Content
	m["Features"] = d.Features
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
