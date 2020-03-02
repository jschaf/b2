package mdext

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/htmls"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

func TestNewImageExt(t *testing.T) {
	const path = "/home/joe/file.md"
	tests := []struct {
		name    string
		src     string
		want    string
		wantCtx map[parser.ContextKey]interface{}
	}{
		{
			"single image",
			texts.Dedent(`
        In a paragraph. ![alt text](./qux.png "title")`),
			texts.Dedent(`
        <p>
          In a paragraph.
          <img src="./qux.png" title="title">
          alt text
        </p>
     `),
			map[parser.ContextKey]interface{}{
				assetsCtxKey: map[string]string{"./qux.png": "/home/joe/qux.png"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewImageExt()))
			buf := new(bytes.Buffer)
			ctx := parser.NewContext()
			SetFilePath(ctx, path)

			if err := md.Convert([]byte(tt.src), buf, parser.WithContext(ctx)); err != nil {
				t.Fatal(err)
			}

			if diff, err := htmls.Diff(buf, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			} else if diff != "" {
				t.Errorf(diff)
			}

			for k, v := range tt.wantCtx {
				if got := ctx.Get(k); !reflect.DeepEqual(got, v) {
					t.Errorf("context key %v, got %s, want %v", k, got, tt.wantCtx[k])
				}
			}
		})
	}
}
