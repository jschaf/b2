import * as md from '//post/mdast/nodes';
import { PostAST } from '//post/post_ast';

describe('PostAST', () => {
  it('should parse a simple mdast', () => {
    const m = md.root([md.paragraph([md.text('hello')])]);

    const p = PostAST.create(m);

    expect(p.mdastNode).toEqual(m);
    expect(p.fnDefsById).toEqual(new Map());
  });

  it('should extract footnote definitions', () => {
    let fnDef1 = md.footnoteDef('1', [md.paragraphText('fn def')]);
    const m = md.root([
      md.paragraph([md.text('hello'), md.footnoteRef('1')]),
      fnDef1,
    ]);

    const p = PostAST.create(m);

    expect(p.mdastNode).toEqual(m);
    expect(p.fnDefsById).toEqual(new Map([['1', fnDef1]]));
  });

  it('should extract inline footnote definitions', () => {
    let inlineFn1 = md.footnote([md.text('inline fn')]);
    let inlineFn2 = md.footnote([md.text('inline fn 2')]);
    const m = md.root([md.paragraph([md.text('hello'), inlineFn1, inlineFn2])]);

    const p = PostAST.create(m);

    expect(p.mdastNode).toEqual(m);
    const id1 = PostAST.newInlineFootnoteId(1);
    const id2 = PostAST.newInlineFootnoteId(2);
    expect(p.fnDefsById).toEqual(
      new Map([
        [id1, md.footnoteDef(id1, [md.paragraph(inlineFn1.children)])],
        [id2, md.footnoteDef(id2, [md.paragraph(inlineFn2.children)])],
      ])
    );
  });
});
