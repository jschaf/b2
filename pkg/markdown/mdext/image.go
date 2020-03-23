package mdext

import (
	"fmt"
	"path"
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
		if !entering {
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() != ast.KindImage {
			return ast.WalkContinue, nil
		}
		img := n.(*ast.Image)

		origDest := string(img.Destination)
		if path.IsAbs(origDest) || strings.HasPrefix(origDest, "http") {
			return ast.WalkContinue, nil
		}
		meta := GetTOMLMeta(pc)
		urlPath := meta.Path
		newDest := path.Join(urlPath, string(img.Destination))
		img.Destination = []byte(newDest)
		sourceDir := filepath.Dir(GetFilePath(pc))
		localPath := filepath.Join(sourceDir, origDest)
		remotePath := filepath.Join(meta.Path, origDest)
		AddAsset(pc, remotePath, localPath)
		return ast.WalkSkipChildren, nil
	})

	if err != nil {
		panic(err)
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
