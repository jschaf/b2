package mdext

import (
	"fmt"
	"github.com/jschaf/b2/pkg/markdown/assets"
	"github.com/jschaf/b2/pkg/markdown/extenders"
	"github.com/jschaf/b2/pkg/markdown/mdctx"
	"github.com/jschaf/b2/pkg/markdown/ord"
	"path"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// imageASTTransformer extracts images we should copy over to the public dir
// when publishing posts.
type imageASTTransformer struct{}

func (f imageASTTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
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
		sourceDir := filepath.Dir(mdctx.GetFilePath(pc))
		localPath := filepath.Join(sourceDir, origDest)
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

// imageRenderer writes images into HTML, replacing the default image renderer.
type imageRenderer struct{}

func (ir imageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, ir.renderImage)
}

func (ir imageRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (status ast.WalkStatus, err error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	tag := fmt.Sprintf(
		"<img src=%q alt=%q title=%q",
		n.Destination, n.Text(source), n.Title)
	_, _ = w.WriteString(tag)
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}
	_, _ = w.WriteString(">")
	return ast.WalkSkipChildren, nil
}

// ImageExt extends markdown with the transformer and renderer.
type ImageExt struct{}

func NewImageExt() *ImageExt {
	return &ImageExt{}
}

func (i *ImageExt) Extend(m goldmark.Markdown) {
	extenders.AddASTTransform(m, imageASTTransformer{}, ord.ImageTransformer)
	extenders.AddRenderer(m, imageRenderer{}, ord.ImageRenderer)
}
