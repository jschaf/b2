import { MdastCompiler } from '//post/mdast/compiler';
import { DeleteCompiler } from '//post/mdast/delete';
import { hastElem, hastElemText, hastText } from '//post/mdast/hast_nodes';
import { PostAST } from '//post/post_ast';
import {
  mdDelete,
  mdEmphasisText,
  mdText,
} from '//post/testing/markdown_nodes';

describe('DeleteCompiler', () => {
  it('should compile a delete', () => {
    const p = PostAST.create(
      mdDelete([mdText('first'), mdEmphasisText('second')])
    );
    const c = MdastCompiler.createDefault();

    const hast = DeleteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      hastElem('del', [hastText('first'), hastElemText('em', 'second')])
    );
  });
});
