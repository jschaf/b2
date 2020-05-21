package mdext

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.uber.org/zap"
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

func (l linkDecorationTransform) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
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

		case strings.HasPrefix(origDest, "https://en.wikipedia.org"):
			link.SetAttribute([]byte("data-link-type"), []byte(linkWiki))
		}

		// If we have a preview, render it into the attributes.
		preview, ok := GetPreview(pc, origDest)
		logger := GetLogger(pc)
		if !ok {
			logger.Debug("no preview for link", zap.String("link", origDest))
			return ast.WalkSkipChildren, nil
		}
		renderer, ok := GetRenderer(pc)
		if !ok {
			panic("link preview: no renderer")
		}

		colonBlock := preview.Parent
		// Assume title is first child.
		title := colonBlock.FirstChild()
		titleHTML := &bytes.Buffer{}
		if err := renderer.Render(titleHTML, title.Text(reader.Source()), title); err != nil {
			panic(fmt.Sprintf("render preview title to HTML for %s: %s", GetFilePath(pc), err.Error()))
		}
		link.SetAttribute([]byte("class"), []byte("preview-target"))
		link.SetAttribute([]byte("data-preview-title"), titleHTML.Bytes())
		link.SetAttribute([]byte("data-preview-snippet"), []byte(texts.Dedent(`
          A <em>snippet</em>. Lorem ipsum dolor <b>sit amet</b>, consectetur 
          adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore 
          magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation 
          ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute 
          <span class="small-caps">IRURE</span> dolor in reprehenderit in voluptate velit esse cillum dolore eu 
          fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, 
          sunt in culpa qui officia deserunt mollit anim id est laborum.`)))

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
