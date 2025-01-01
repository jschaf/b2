package mdext

import (
	"testing"

	"github.com/jschaf/jsc/pkg/markdown/mdtest"

	"github.com/jschaf/jsc/pkg/texts"
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
		    <img src="qux.png" loading="lazy" alt="alt text" title="title">
		  </picture>
			  </figure>
		`),
		},
		{
			"single image with caption",
			texts.Dedent(`
		  ![alt text](./bar.png "title")
		
		  CAPTION: foobar
		`),
			texts.Dedent(`
			  <figure>
					<picture>
						<img src="bar.png" loading="lazy" alt="alt text" title="title">
					</picture>
					<figcaption>
						<span class="caption-label">Figure:</span>
						foobar
					</figcaption>
			  </figure>
		`),
		},
		{
			"single relative image with caption with slug",
			texts.Dedent(`
      +++
      slug = "some_slug"
      +++

		  ![alt text](./bar.png "title")
		
		  CAPTION: foobar
		`),
			texts.Dedent(`
			  <figure>
					<picture>
						<img src="/some_slug/bar.png" loading="lazy" alt="alt text" title="title">
					</picture>
					<figcaption>
						<span class="caption-label">Figure:</span>
						foobar
					</figcaption>
			  </figure>
		`),
		},
		{
			"single absolute image with caption with slug",
			texts.Dedent(`
      +++
      slug = "some_slug"
      +++

		  ![alt text](https://example.com/bar.png "title")
		
		  CAPTION: foobar
		`),
			texts.Dedent(`
			  <figure>
					<picture>
						<img src="https://example.com/bar.png" loading="lazy" alt="alt text" title="title">
					</picture>
					<figcaption>
						<span class="caption-label">Figure:</span>
						foobar
					</figcaption>
			  </figure>
		`),
		},
		{
			"single image nested in list",
			texts.Dedent(`
      +++
      slug = "some_slug"
      +++
      
      1. one

         ![alt text](https://example.com/bar.png "title")

         CAPTION: foobar
      2. two
		`),
			texts.Dedent(`
        <ol>
          <li>
            <p>one</p>
						<figure>
							<picture>
								<img src="https://example.com/bar.png" loading="lazy" alt="alt text" title="title">
							</picture>
							<figcaption>
								<span class="caption-label">Figure:</span>
								foobar
							</figcaption>
						</figure>
          </li>
          <li>
            <p>two</p>
          </li>
        </ol>
		`),
		},
		{
			"complex image with caption",
			texts.Dedent(`
        foo bar

        ![alt text](bar.png "title")

        CAPTION: foobar
     `),
			texts.Dedent(`
        <p>
          foo bar
        </p>
			  <figure>
					<picture>
						<img src="bar.png" loading="lazy" alt="alt text" title="title">
					</picture>
					<figcaption>
						<span class="caption-label">Figure:</span>
						foobar
					</figcaption>
			  </figure>
    `),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewTOMLExt(), NewFigureExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
