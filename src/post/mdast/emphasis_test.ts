import { MdastCompiler } from '//post/mdast/compiler';
import { EmphasisCompiler } from '//post/mdast/emphasis';
import { hastElem, hastText } from '//post/mdast/hast_nodes';
import { PostAST } from '//post/post_ast';
import { mdEmphasisText } from '//post/testing/markdown_nodes';

describe('EmphasisCompiler', () => {
  it('should compile emphasis with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(mdEmphasisText(content));
    const c = MdastCompiler.createDefault();

    const hast = EmphasisCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('em', [hastText(content)]));
  });
});
