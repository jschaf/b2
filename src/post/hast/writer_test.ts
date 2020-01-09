import { PostAST } from '//post/ast';
import { HastWriter } from '//post/hast/writer';
import * as h from '//post/hast/nodes';
import * as md from '//post/mdast/nodes';

describe('HastWriter', () => {
  it('should compile body > p', () => {
    const ast = PostAST.fromMdast(md.root([]));
    const a = h.elem('body', [h.elemText('p', 'foo bar')]);

    const html = HastWriter.createDefault().write(a, ast);

    expect(html).toMatchSnapshot();
  });
});
