package html

import (
	"fmt"
	"html/template"
	"reflect"
)

var fns = template.FuncMap{
	"isLast": isLast,
}

type PostTemplateData struct {
	Title string
	Body  template.HTML
}

var PostTemplate = template.Must(
	template.New("post_template.gohtml").Funcs(fns).ParseFiles(
		"pkg/markdown/html/post_template.gohtml"))

type IndexTemplateData struct {
	Title  string
	Bodies []template.HTML
}

var IndexTemplate = template.Must(
	template.New("index_template.gohtml").Funcs(fns).ParseFiles(
		"pkg/markdown/html/index_template.gohtml"))

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
