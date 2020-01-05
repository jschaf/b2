import { PostCompiler } from '//post/compiler';
import { PostAST } from '//post/post_ast';
import { PostBag } from '//post/post_bag';
import { withDefaultFrontMatter } from '//post/testing/front_matters';
import { dedent } from '//strings';

describe('PostCompiler', () => {
  it('should compile a simple post', async () => {
    const md = withDefaultFrontMatter(dedent`
      # hello
      
      Hello world.
    `);
    const bag = PostBag.fromMarkdown(md);
    const ast = PostAST.fromMdast(bag.postNode.node);

    const actual = PostCompiler.create().compileToMempost(ast);

    expect(actual).toEqualMempost({
      'index.html': `
        <html>
        <head></head>
        <h1>hello</h1>
        <p>Hello world.</p>
        </html>
      `,
    });
  });
});
