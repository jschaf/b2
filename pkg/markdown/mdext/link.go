package mdext

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type LinkTransformer struct{}

type linkType = string

const (
	linkPDF  linkType = "pdf"
	linkWiki linkType = "wikipedia"
)

func (l *LinkTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() != ast.KindLink {
			return ast.WalkContinue, nil
		}

		link := n.(*ast.Link)
		origDest := string(link.Destination)

		if filepath.IsAbs(origDest) || strings.HasPrefix(origDest, "http") {
			return ast.WalkContinue, nil
		}
		filePath := filepath.Dir(GetFilePath(pc))
		meta := GetTOMLMeta(pc)
		newDest := path.Join(meta.Path, origDest)
		link.Destination = []byte(newDest)
		localPath := filepath.Join(filePath, origDest)
		remotePath := filepath.Join(meta.Path, origDest)
		AddAsset(pc, remotePath, localPath)

		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		panic(err)
	}
}

type linkDecorationTransform struct{}

func (l linkDecorationTransform) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() != ast.KindLink {
			return ast.WalkContinue, nil
		}

		link := n.(*ast.Link)
		origDest := string(link.Destination)

		switch {
		case path.Ext(origDest) == ".pdf":
			link.SetAttribute([]byte("data-link-type"), []byte(linkPDF))
			link.SetAttribute([]byte("class"), []byte("preview-target"))

		case strings.HasPrefix(origDest, "https://en.wikipedia.org"):
			link.SetAttribute([]byte("data-link-type"), []byte(linkWiki))
			link.SetAttribute([]byte("class"), []byte("preview-target"))
		}
		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		panic(err)
	}
}

type LinkExt struct{}

func NewLinkExt() *LinkExt {
	return &LinkExt{}
}

func (l *LinkExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&linkDecorationTransform{}, 900),
			util.Prioritized(&LinkTransformer{}, 901)))
}
