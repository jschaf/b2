package mdext

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/htmls"
	"github.com/jschaf/b2/pkg/texts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

func TestNewFigureExt(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"single image",
			texts.Dedent(`
        ![alt text](./qux.png "title")`),
			texts.Dedent(`
			  <figure>
         <picture>
           <img src="./qux.png" title="title">
         </picture>
			  </figure>
     `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(
				NewFigureExt()))
			buf := new(bytes.Buffer)
			ctx := parser.NewContext()

			if err := md.Convert([]byte(tt.src), buf, parser.WithContext(ctx)); err != nil {
				t.Fatal(err)
			}

			if diff, err := htmls.Diff(buf, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			} else if diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
