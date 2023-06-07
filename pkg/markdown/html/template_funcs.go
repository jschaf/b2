package html

import (
	"html/template"
	"time"
)

func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"now":    time.Now,
		"isLast": isLast,
	}
}
