package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/jschaf/b2/pkg/htmls"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/jschaf/b2/pkg/texts"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"h1 > p",
			withFrontmatter(mdext.PostMeta{Slug: "foo"},
				`
        # hello world
        para
      `),
			articleHTML(mdext.PostMeta{Slug: "foo"}, "hello world", "<p>para</p>"),
		},
		{
			"img - has asset",
			withFrontmatter(mdext.PostMeta{Slug: "foo"},
				`
        # title
        ![Alt text](./foo_bar)
      `),
			articleHTML(mdext.PostMeta{Slug: "foo"}, "title",
				"<figure><picture><img src=\"./foo_bar\" title=\"\"></picture></figure>",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := New()
			ast, err := md.Parse("", strings.NewReader(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			got := new(bytes.Buffer)
			if err := md.Render(got, []byte(tt.input), ast); err != nil {
				t.Fatal(err)
			}

			if diff, err := htmls.DiffStrings(tt.want, got.String()); err != nil {
				t.Error(err)
			} else if diff != "" {
				t.Errorf("Render() got:\n%s\nwant:\n%s", got.String(), tt.want)
			}
		})
	}
}

func withFrontmatter(meta mdext.PostMeta, md string) string {
	b := new(bytes.Buffer)
	b.WriteString("+++\n")
	if meta.Slug != "" {
		b.WriteString("slug = \"" + meta.Slug + "\"\n")
	}
	b.WriteString("date = " + meta.Date.Format("2006-01-02") + "\n")
	b.WriteString("+++\n\n")
	b.WriteString(texts.Dedent(md))
	return b.String()
}

func articleHTML(meta mdext.PostMeta, title, text string) string {
	b := new(bytes.Buffer)
	b.WriteString("<article>\n")
	b.WriteString("<header>\n")
	b.WriteString(fmt.Sprintf(
		"<time datetime=%q>%s</time>\n",
		meta.Date.UTC().Format("2006-01-02"),
		meta.Date.Format("January _2, 2006"),
	))
	b.WriteString("<h1>")
	b.WriteString(fmt.Sprintf("<a href=%q title=%q>", "/"+meta.Slug, title))
	b.WriteString(title)
	b.WriteString("</a>")
	b.WriteString("</h1>\n")
	b.WriteString("</header>\n")
	b.WriteString(text + "\n")
	b.WriteString("</article>\n")
	return b.String()
}
