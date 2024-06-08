package mdext

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/ord"

	"github.com/jschaf/b2/pkg/markdown/asts"
	"github.com/jschaf/b2/pkg/markdown/attrs"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// linkAssetTransformer is an AST transformer to extract assets that need to be
// copied to the serving directory like local links to images or PDFs.
type linkAssetTransformer struct{}

type linkType = string

func (l *linkAssetTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
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
		filePath := filepath.Dir(mdctx.GetFilePath(pc))
		meta := GetTOMLMeta(pc)
		newDest := path.Join(meta.Path, origDest)
		link.Destination = []byte(newDest)
		localPath := filepath.Join(filePath, origDest)
		remotePath := filepath.Join(meta.Path, origDest)
		mdctx.AddAsset(pc, assets.Blob{
			Src:  localPath,
			Dest: remotePath,
		})

		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		panic(err)
	}
}

// linkDecorationTransform is an AST transformer that adds preview information
// to links.
type linkDecorationTransform struct{}

const (
	LinkCitation linkType = "citation"
	LinkPDF      linkType = "pdf"
	LinkWiki     linkType = "wikipedia"
)

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
			link.SetAttribute([]byte("data-link-type"), []byte(LinkPDF))

		case strings.HasPrefix(origDest, "https://en.wikipedia.org"):
			link.SetAttribute([]byte("data-link-type"), []byte(LinkWiki))
		}

		renderPreview(pc, origDest, reader, link)

		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		panic(err)
	}
}

func renderPreview(pc parser.Context, origDest string, reader text.Reader, link *ast.Link) {
	// If we have a preview, render it into the attributes.
	preview, ok := GetPreview(pc, origDest)
	if !ok {
		return
	}
	renderer, ok := mdctx.GetRenderer(pc)
	if !ok {
		panic("link preview: no renderer")
	}

	colonBlock := preview.Parent
	if colonBlock == nil {
		return
	}
	// Assume title is first child.
	title := colonBlock.FirstChild()
	if title == nil {
		return
	}
	attrs.AddClass(title, "preview-title")
	titleLink := ast.NewLink()
	titleLink.Destination = []byte(origDest)
	asts.Reparent(titleLink, title)
	title.AppendChild(title, titleLink)
	title.SetAttributeString(attrs.CustomTagAttr, "div")
	titleHTML := &bytes.Buffer{}
	if err := renderer.Render(titleHTML, reader.Source(), title); err != nil {
		panic(fmt.Sprintf("render preview title to HTML for %s: %s", mdctx.GetFilePath(pc), err.Error()))
	}
	link.SetAttribute([]byte("class"), []byte("preview-target"))
	link.SetAttribute([]byte("data-preview-title"), bytes.Trim(titleHTML.Bytes(), " \n"))

	// Assume the rest of the children are the body.
	snippetHTML := &bytes.Buffer{}
	snippetNode := title.NextSibling()
	for snippetNode != nil {
		if err := renderer.Render(snippetHTML, reader.Source(), snippetNode); err != nil {
			panic(fmt.Sprintf("render preview snippet to HTML for %s: %s", mdctx.GetFilePath(pc), err.Error()))
		}
		snippetNode = snippetNode.NextSibling()
	}
	link.SetAttribute([]byte("data-preview-snippet"), bytes.Trim(snippetHTML.Bytes(), " \n"))
}

type LinkExt struct{}

func NewLinkExt() *LinkExt {
	return &LinkExt{}
}

func (l *LinkExt) Extend(m goldmark.Markdown) {
	extenders.AddASTTransform(m, &linkDecorationTransform{}, ord.LinkDecorationTransformer)
	extenders.AddASTTransform(m, &linkAssetTransformer{}, ord.LinkAssetTransformer)
}
