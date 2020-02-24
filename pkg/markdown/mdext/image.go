package mdext

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ImageASTTransformer extracts images we should copy over to the public dir.
type ImageASTTransformer struct{}

func (f *ImageASTTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() != ast.KindImage {
			return ast.WalkContinue, nil
		}
		img := n.(*ast.Image)

		dest := string(img.Destination)
		if filepath.IsAbs(dest) || strings.HasPrefix(dest, "http") {
			return ast.WalkContinue, nil
		}
		path := filepath.Dir(GetPath(pc))
		meta := GetTOMLMeta(pc)
		localPath := filepath.Join(path, dest)
		remotePath := filepath.Join(meta.Slug, dest)
		AddAsset(pc, remotePath, localPath)
		return ast.WalkStop, nil
	})

	if err != nil {
		log.Printf("error in image AST transformer: %s", err)
	}
}

type ImageRenderer struct{}

func (i *ImageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, i.renderImage)
}

func (i *ImageRenderer) renderImage(w util.BufWriter, _ []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString(fmt.Sprintf("<img src=%q title=%q>", n.Destination, n.Title))
	return ast.WalkContinue, nil
}

type ImageExt struct{}

func (i *ImageExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&ImageASTTransformer{}, 999)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&ImageRenderer{}, 500),
	))
}

func NewImageExt() *ImageExt {
	return &ImageExt{}
}
