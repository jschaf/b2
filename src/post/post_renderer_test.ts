import { withDefaultFrontMatter } from '//post/testing/front_matters';
import { dedent } from '//strings';
import { PostBag } from '//post/post_bag';
import { PostRenderer } from '//post/post_renderer';

describe('PostRenderer', () => {
  it('should render a simple post', async () => {
    const md = withDefaultFrontMatter(dedent`
      # hello
      
      Hello world.
    `);
    const bag = PostBag.fromMarkdown(md);

    const actual = await PostRenderer.create().render(bag);

    expect(actual).toEqualMempost({
      'index.html': `
        <h1>hello</h1>
        <p>Hello world.</p>
      `,
    });
  });
});
