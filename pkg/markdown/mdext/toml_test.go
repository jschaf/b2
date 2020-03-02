package mdext

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

func TestMeta(t *testing.T) {
	source := `+++
slug = "a slug"
date = 2019-09-20
+++
# Hello goldmark-meta
`

	md := goldmark.New(goldmark.WithExtensions(NewTOMLExt()))
	var buf bytes.Buffer
	context := parser.NewContext()
	if err := md.Convert([]byte(source), &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	meta := GetTOMLMeta(context)
	if meta.Slug != "a slug" {
		t.Errorf("Title must be %s, but got %v", "a slub", meta.Slug)
	}
	if buf.String() != "<h1>Hello goldmark-meta</h1>\n" {
		t.Errorf("should renderFigure '<h1>Hello goldmark-meta</h1>', but '%s'", buf.String())
	}
}
