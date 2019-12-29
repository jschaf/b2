import { withDefaultFrontMatter } from '//post/testing/front_matters';
import { dedent } from '//strings';
import { PostBag } from '//post/post_bag';
import { PostHtmlRenderer } from '//post/render_html/render';

describe('PostHtmlRenderer', () => {
  it('should render a simple post', async () => {
    const md = withDefaultFrontMatter(dedent`
      # hello
      
      Hello world.
    `);
    const bag = PostBag.fromMarkdown(md);

    const actual = await PostHtmlRenderer.create().render(bag);

    expect(actual).toEqualMempost({
      'index.html': `
        <!doctype html>
        <html lang="en">
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1">
        </head>
        <h1>hello</h1>
        <p>Hello world.</p>
        </html>
      `,
    });
  });
});
