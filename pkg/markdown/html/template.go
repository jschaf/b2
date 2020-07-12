package html

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"html/template"
	"io"
	"path/filepath"
	"reflect"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/js"
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
	layouts := []string{"index.gohtml", "post.gohtml"}
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
	result, err := js.BundleMain()
	if err != nil {
		return fmt.Errorf("failed to bundle main.js: %w", err)
	}
	data["SyncScript"] = template.JS(result.JsContents)

	return tmpl.ExecuteTemplate(w, "base", data)
}

func RenderPost(w io.Writer, d PostTemplateData) error {
	m := make(map[string]interface{})
	m["Title"] = d.Title
	m["Content"] = d.Content
	return render(w, "post.gohtml", m)
}

func RenderIndex(w io.Writer, d IndexTemplateData) error {
	m := make(map[string]interface{})
	m["Title"] = d.Title
	m["Bodies"] = d.Bodies
	return render(w, "index.gohtml", m)
}

type MainTemplateData struct {
	Title   string
	Content template.HTML
}

type PostTemplateData struct {
	Title   string
	Content template.HTML
}

type IndexTemplateData struct {
	Title  string
	Bodies []template.HTML
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
