import * as md from '//post/mdast/nodes';
import { PostAST } from '//post/post_ast';
import * as mdast from 'mdast';

describe('PostAST', () => {
  it('should parse a simple mdast', () => {
    const m = md.root([md.paragraph([md.text('hello')])]);

    const p = PostAST.create(m);

    expect(p.mdastNode).toEqual(m);
    expect(p.fnDefsById).toEqual(new Map());
  });

  describe('link definition extraction', () => {
    const defA = md.definitionProps('a', 'a-url', { title: 'a-title' });
    const defB = md.definition('b', 'b-url');
    const defC = md.definition('  \t\nc  c\n\t', 'c-url');
    const defD = md.definition('D  d', 'd-url');

    type LinkDefEntries = [string, mdast.Definition][];
    const linkDefTests: [string, mdast.Content[], LinkDefEntries][] = [
      ['1 def', [defA], [['a', defA]]],
      [
        '2 defs',
        [defA, defB],
        [
          ['a', defA],
          ['b', defB],
        ],
      ],
      ['whitespace', [defC], [['c c', defC]]],
      ['case folding', [defD], [['d d', defD]]],
    ];
    for (const [name, children, expected] of linkDefTests) {
      it(`should handle ${name}`, () => {
        const p = PostAST.create(md.root(children));
        expect(p.defsById).toEqual(new Map(expected));
      });

      it(`should be able to getDefinition for ${name}`, () => {
        const p = PostAST.create(md.root(children));
        for (const [key, def] of expected) {
          expect(p.getDefinition(key)).toEqual(def);
          expect(p.getDefinition(key.toUpperCase())).toEqual(def);
          expect(p.getDefinition(`  \t${key}  `)).toEqual(def);
        }
      });
    }
  });

  describe('footnote definition extraction', () => {
    const defA = md.footnoteDef('1', [md.paragraphText('a-def')]);
    const defB = md.footnoteDef('b', [md.paragraphText('b-def')]);
    const defC = md.footnoteDef('\n  \tc \tc \t', [md.paragraphText('c-def')]);
    const defD = md.footnoteDef('D d', [md.paragraphText('d-def')]);

    const id1 = PostAST.newInlineFootnoteId(1);
    const id2 = PostAST.newInlineFootnoteId(2);
    const inA = md.footnote([md.text('a-def')]);
    const inDefA = md.footnoteDef(id1, [md.paragraph(inA.children)]);
    const inB = md.footnote([md.text('b-def')]);
    const inDefB = md.footnoteDef(id2, [md.paragraph(inB.children)]);

    type FnDefEntries = [string, mdast.FootnoteDefinition][];
    const tests: [string, mdast.Content[], FnDefEntries][] = [
      ['1 def', [defA], [['1', defA]]],
      [
        '2 defs',
        [defA, defB],
        [
          ['1', defA],
          ['b', defB],
        ],
      ],
      ['whitespace', [defC], [['c c', defC]]],
      ['case folding', [defD], [['d d', defD]]],

      ['1 inline def', [inA], [[id1, inDefA]]],
      [
        '2 inline defs',
        [inA, inB],
        [
          [id1, inDefA],
          [id2, inDefB],
        ],
      ],
      [
        'mixed inline',
        [inA, md.html('<br>'), inB],
        [
          [id1, inDefA],
          [id2, inDefB],
        ],
      ],
    ];
    for (const [name, children, expected] of tests) {
      it(`should handle ${name}`, () => {
        const p = PostAST.create(md.root(children));
        expect(p.fnDefsById).toEqual(new Map(expected));
      });

      it(`should be able to getFnDef for ${name}`, () => {
        const p = PostAST.create(md.root(children));
        for (const [key, def] of expected) {
          expect(p.getFootnoteDef(key)).toEqual(def);
        }
      });
    }
  });
});
