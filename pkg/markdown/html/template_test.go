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
	err := RenderPostDetail(w, PostDetailData{
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
	body1 := "<div>body1</div>"
	body2 := "<div>body2</div>"
	data := RootIndexData{
		Title:    title,
		Bodies:   []template.HTML{template.HTML(body1), template.HTML(body2)},
		Features: mdctx.NewFeatureSet(),
	}

	if err := RenderRootIndex(w, data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.String(), title) {
		t.Errorf("rendered content doesn't include %q:\n\n%s", title, w.String())
	}
	if !strings.Contains(w.String(), body1) {
		t.Errorf("rendered content doesn't include %q:\n\n%s", body1, w.String())
	}
	if !strings.Contains(w.String(), body2) {
		t.Errorf("rendered content doesn't include %q:\n\n%s", body2, w.String())
	}
}
