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

func TestNewLinkExt(t *testing.T) {
	const path = "/home/joe/file.md"
	tests := []struct {
		name    string
		src     string
		want    string
		wantCtx map[parser.ContextKey]interface{}
	}{
		{
			"single relative link",
			texts.Dedent(`
				Paper: [Gorilla Title][gorilla]
		
				[gorilla]: paper.pdf
     `),
			texts.Dedent(`
       <p>
         Paper: <a href="paper.pdf">Gorilla Title</a>
       </p>
    `),
			map[parser.ContextKey]interface{}{
				assetsCtxKey: map[string]string{"paper.pdf": "/home/joe/paper.pdf"},
			},
		},
		{
			"single relative link with slug",
			texts.Dedent(`
				+++
				slug = "some_slug"
				+++
		
				Paper: [Gorilla Title][gorilla]
		
				[gorilla]: paper.pdf
     `),
			texts.Dedent(`
       <p>
         Paper: <a href="/some_slug/paper.pdf">Gorilla Title</a>
       </p>
    `),
			map[parser.ContextKey]interface{}{
				assetsCtxKey: map[string]string{"/some_slug/paper.pdf": "/home/joe/paper.pdf"},
			},
		},
		{
			"single absolute link with slug",
			texts.Dedent(`
				+++
				slug = "some_slug"
				+++

				Paper: [Gorilla Title][gorilla]

				[gorilla]: http://example.com/paper.pdf
      `),
			texts.Dedent(`
        <p>
          Paper: <a href="http://example.com/paper.pdf">Gorilla Title</a>
        </p>
     `),
			map[parser.ContextKey]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewTOMLExt(),
				NewLinkExt()))
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
