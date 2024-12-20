package html

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/markdown/mdctx"
)

func TestRenderPost(t *testing.T) {
	w := &bytes.Buffer{}
	title := "foo_title"
	content := "<b>foo_content</b>"
	err := RenderDetail(w, DetailParams{
		Title:    title,
		Content:  template.HTML(content),
		Features: mdctx.NewFeatureSet(),
	})
	if err := err; err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.String(), title) {
		t.Errorf("rendered content doesn't include %q:\n\n%s", title, w.String())
	}
	if !strings.Contains(w.String(), content) {
		t.Errorf("rendered content doesn't include %q:\n\n%s", content, w.String())
	}
}

func TestRenderIndex(t *testing.T) {
	w := &bytes.Buffer{}
	title := "foo_title"
	data := IndexParams{
		Title: title,
		Posts: []IndexPostData{
			{Title: "post1", Body: template.HTML("body")},
			{Title: "post2", Body: template.HTML("body")},
		},
		Features: mdctx.NewFeatureSet(),
	}

	if err := RenderIndex(w, data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.String(), title) {
		t.Errorf("rendered content doesn't include %q:\n\n%s", title, w.String())
	}
	if !strings.Contains(w.String(), "post1") {
		t.Errorf("rendered content doesn't include %q:\n\n%s", "post1", w.String())
	}
	if !strings.Contains(w.String(), "post2") {
		t.Errorf("rendered content doesn't include %q:\n\n%s", "post2", w.String())
	}
}
