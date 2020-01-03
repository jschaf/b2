import { MdastCompiler } from '//post/mdast/compiler';
import { FootnoteCompiler } from '//post/mdast/footnote';
import { FootnoteReferenceCompiler } from '//post/mdast/footnote_reference';
import { PostAST } from '//post/post_ast';
import { mdInlineFootnote, mdText } from '//post/testing/markdown_nodes';

describe('FootnoteCompiler', () => {
  it('should compile a footnote', () => {
    const p = PostAST.create(mdInlineFootnote([mdText('inline fn')]));
    const c = MdastCompiler.createDefault();

    const hast = FootnoteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      FootnoteReferenceCompiler.makeHastNode(PostAST.newInlineFootnoteId(1))
    );
  });
});
