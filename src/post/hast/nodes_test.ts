import { danglingImageRef, danglingLinkRef } from '//post/hast/nodes';
import { MdastCompiler } from '//post/mdast/compiler';
import * as md from '//post/mdast/nodes';
import * as h from '//post/hast/nodes';
import { PostAST } from '//post/post_ast';
import * as unist from 'unist';
import * as mdast from 'mdast';

describe('danglingImageRef', () => {
  const attrs: [md.RefType, string, md.ImageRefProps, string][] = [
    [md.RefType.Full, 'alpha', { alt: 'alt' }, '![alt][alpha]'],
    [md.RefType.Full, 'alpha', { alt: 'alt', label: 'ALPHA' }, '![alt][ALPHA]'],
    [md.RefType.Shortcut, 'alpha', {}, '![alpha]'],
    [md.RefType.Shortcut, 'alpha', { label: 'ALPHA' }, '![ALPHA]'],
    [md.RefType.Collapsed, 'alpha', {}, '![alpha][]'],
    [md.RefType.Collapsed, 'alpha', { label: 'ALPHA' }, '![ALPHA][]'],
  ];
  for (let [ref, id, props, expected] of attrs) {
    it(`should render ref=${ref}, label=${props.label || '<none>'}`, () => {
      const ir = md.imageRefProps(id, ref, props);
      expect(danglingImageRef(ir)).toEqual(h.text(expected));
    });
  }
});

describe('danglingLinkRef', () => {
  const full = md.RefType.Full;
  const shortcut = md.RefType.Shortcut;
  const collapsed = md.RefType.Collapsed;
  const attrs: [mdast.LinkReference, unist.Node[]][] = [
    [md.linkRefText('id', full, 'foo'), [h.text('[foo][id]')]],
    [
      md.linkRefProps('id', full, { label: 'ID' }, [md.text('foo')]),
      [h.text('[foo][ID]')],
    ],
    [md.linkRefText('id', shortcut, 'foo'), [h.text('[foo]')]],
    [md.linkRefText('id', collapsed, 'foo'), [h.text('[foo][]')]],
    [
      md.linkRef('id', collapsed, [md.emphasisText('a'), h.text('b')]),
      [h.text('['), h.elemText('em', 'a'), h.text('b][]')],
    ],
  ];
  const c = MdastCompiler.createDefault();
  const compileChildren = (n: unist.Parent) =>
    c.compileChildren(n, PostAST.fromMdast(n));

  for (const [lr, expected] of attrs) {
    const l = lr.label || '<none>';
    const id = lr.identifier;
    it(`should render ref=${lr.referenceType}, id=${id}, label=${l}`, () => {
      expect(danglingLinkRef(lr, compileChildren)).toEqual(expected);
    });
  }
});
