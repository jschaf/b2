package html

import "html/template"

type TemplateData struct {
	Title string
	Body  template.HTML
}

var PostDoc = template.Must(template.ParseFiles("pkg/html/post_doc.gohtml"))
