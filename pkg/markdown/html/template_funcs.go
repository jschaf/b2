package html

import (
	"fmt"
	"html/template"
	"reflect"
	"time"
)

func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"now":    time.Now,
		"isLast": isLast,
	}
}

// isLast returns true if the index is the last index in the item.
func isLast(index int, item any) (bool, error) {
	v := reflect.ValueOf(item)
	if !v.IsValid() {
		return false, fmt.Errorf("isLast of untyped nil")
	}
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return index == v.Len()-1, nil
	default:
		return false, fmt.Errorf("isLast of type %s", v.Type())
	}
}
