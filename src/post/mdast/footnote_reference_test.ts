import { FootnoteReferenceCompiler } from '//post/mdast/footnote_reference';
import { PostAST } from '//post/post_ast';
import { mdFootnoteRef } from '//post/testing/markdown_nodes';

describe('FootnoteReferenceCompiler', () => {
  it('should compile a footnote reference', () => {
    const id = 'my-fn-ref';
    const p = PostAST.create(mdFootnoteRef(id));

    const hast = FootnoteReferenceCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(FootnoteReferenceCompiler.makeHastNode(id));
  });
});
