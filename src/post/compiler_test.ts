import { PostCompiler } from '//post/compiler';
import { PostAST } from '//post/ast';
import * as frontMatters from '//post/testing/front_matters';
import * as md from '//post/mdast/nodes';

describe('PostCompiler', () => {
  it('should compile a simple post', async () => {
    const ast = PostAST.fromMdast(
      md.root([
        frontMatters.defaultTomlMdast(),
        md.headingText('h1', 'alpha'),
        md.paragraphText('Foo bar.'),
      ])
    );

    const actual = PostCompiler.create().compile(ast);

    expect(actual.mempost).toEqualMempost({
      'index.html': `
        <html>
        <head></head>
        <h1>alpha</h1>
        <p>Foo bar.</p>
        </html>
      `,
    });
  });
});
