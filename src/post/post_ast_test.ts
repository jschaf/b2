import { PostAST } from '//post/post_ast';
import {
  mdInlineFootnote,
  mdFootnoteDef,
  mdFootnoteRef,
  mdPara,
  mdParaText,
  mdRoot,
  mdText,
} from '//post/testing/markdown_nodes';

describe('PostAST', () => {
  it('should parse a simple mdast', () => {
    const md = mdRoot([mdPara([mdText('hello')])]);

    const ast = PostAST.create(md);

    expect(ast.mdastNode).toEqual(md);
    expect(ast.fnDefsById).toEqual(new Map());
  });

  it('should extract footnote definitions', () => {
    let fnDef1 = mdFootnoteDef('1', [mdParaText('fn def')]);
    const md = mdRoot([mdPara([mdText('hello'), mdFootnoteRef('1')]), fnDef1]);

    const ast = PostAST.create(md);

    expect(ast.mdastNode).toEqual(md);
    expect(ast.fnDefsById).toEqual(new Map([['1', fnDef1]]));
  });

  it('should extract inline footnote definitions', () => {
    let inlineFn1 = mdInlineFootnote([mdText('inline fn')]);
    let inlineFn2 = mdInlineFootnote([mdText('inline fn 2')]);
    const md = mdRoot([mdPara([mdText('hello'), inlineFn1, inlineFn2])]);

    const ast = PostAST.create(md);

    expect(ast.mdastNode).toEqual(md);
    const id1 = PostAST.newInlineFootnoteId(1);
    const id2 = PostAST.newInlineFootnoteId(2);
    expect(ast.fnDefsById).toEqual(
      new Map([
        [id1, mdFootnoteDef(id1, [mdPara(inlineFn1.children)])],
        [id2, mdFootnoteDef(id2, [mdPara(inlineFn2.children)])],
      ])
    );
  });
});
