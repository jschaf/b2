import { PostAST } from '//post/ast';
import { HastCompiler } from '//post/hast/compiler';
import * as h from '//post/hast/nodes';
import * as md from '//post/mdast/nodes';

describe('HastCompiler', () => {
  it('should compile body > p', () => {
    const ast = PostAST.fromMdast(md.root([]));
    const a = h.elem('body', [h.elemText('p', 'foo bar')]);

    const html = HastCompiler.create().compile(a, ast);

    expect(html).toEqual('<body><p>foo bar</p></body>');
  });
});
