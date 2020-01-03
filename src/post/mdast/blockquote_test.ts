import { BlockquoteCompiler } from '//post/mdast/blockquote';
import { MdastCompiler } from '//post/mdast/compiler';
import { hastElem, hastElemText } from '//post/mdast/hast_nodes';
import { PostAST } from '//post/post_ast';
import {
  mdBlockquote,
  mdEmphasisText,
  mdPara,
  mdParaText,
} from '//post/testing/markdown_nodes';

describe('BlockquoteCompiler', () => {
  it('should compile a blockquote', () => {
    const p = PostAST.create(
      mdBlockquote([mdParaText('first'), mdPara([mdEmphasisText('second')])])
    );
    const c = MdastCompiler.createDefault();

    const hast = BlockquoteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      hastElem('blockquote', [
        hastElemText('p', 'first'),
        hastElem('p', [hastElemText('em', 'second')]),
      ])
    );
  });
});
