import { BreakCompiler } from '//post/mdast/break';
import { hastElem } from '//post/mdast/hast_nodes';
import { PostAST } from '//post/post_ast';
import { mdBreak } from '//post/testing/markdown_nodes';

describe('BreakCompiler', () => {
  it('should compile a break', () => {
    const p = PostAST.create(mdBreak());

    const hast = BreakCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('break'));
  });
});
